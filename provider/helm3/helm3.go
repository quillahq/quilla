package helm3

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/quilla-hq/quilla/approvals"
	"github.com/quilla-hq/quilla/internal/policy"
	"github.com/quilla-hq/quilla/types"
	"github.com/quilla-hq/quilla/util/image"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/quilla-hq/quilla/extension/notification"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"helm.sh/helm/v3/pkg/strvals"

	_ "helm.sh/helm/v3/pkg/action"
	_ "helm.sh/helm/v3/pkg/release"

	hapi_chart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

var helm3VersionedUpdatesCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "helm3_versioned_updates_total",
		Help: "How many versioned helm3 charts were updated, partitioned by chart name.",
	},
	[]string{"chart"},
)

var helm3UnversionedUpdatesCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "helm3_unversioned_updates_total",
		Help: "How many unversioned helm3 charts were updated, partitioned by chart name.",
	},
	[]string{"chart"},
)

func init() {
	prometheus.MustRegister(helm3VersionedUpdatesCounter)
	prometheus.MustRegister(helm3UnversionedUpdatesCounter)
}

// ErrPolicyNotSpecified helm related errors
var (
	ErrPolicyNotSpecified = errors.New("policy not specified")
)

// Manager - high level interface into helm provider related data used by
// triggers
type Manager interface {
	Images() ([]*image.Reference, error)
}

// ProviderName - helm provider name
const ProviderName = "helm3"

// DefaultUpdateTimeout - update timeout in seconds
// const DefaultUpdateTimeout = 300

// UpdatePlan - release update plan
type UpdatePlan struct {
	Namespace string
	Name      string

	Config *quillaChartConfig

	// chart
	Chart *hapi_chart.Chart

	// values to update path=value
	Values map[string]string

	// Current (last seen cluster version)
	CurrentVersion string
	// New version that's already in the deployment
	NewVersion string

	// ReleaseNotes is a slice of combined release notes.
	ReleaseNotes []string

	// used as fix to bug in chartutil.coalesce v3.1.2
	EmptyConfig bool
}

// quilla:
//   # quilla policy (all/major/minor/patch/force)
//   policy: all
//   # trigger type, defaults to events such as pubsub, webhooks
//   trigger: poll
//   pollSchedule: "@every 2m"
//   # images to track and update
//   images:
//     - repository: image.repository
//       tag: image.tag

// Root - root element of the values yaml
type Root struct {
	quilla quillaChartConfig `json:"quilla"`
}

// quillaChartConfig - quilla related configuration taken from values.yaml
type quillaChartConfig struct {
	Policy               string            `json:"policy"`
	MatchTag             bool              `json:"matchTag"`
	MatchPreRelease      bool              `json:"matchPreRelease"`
	Trigger              types.TriggerType `json:"trigger"`
	PollSchedule         string            `json:"pollSchedule"`
	Approvals            int               `json:"approvals"`        // Minimum required approvals
	ApprovalDeadline     int               `json:"approvalDeadline"` // Deadline in hours
	Images               []ImageDetails    `json:"images"`
	NotificationChannels []string          `json:"notificationChannels"` // optional notification channels

	Plc policy.Policy `json:"-"`
}

// ImageDetails - image details
type ImageDetails struct {
	RepositoryPath  string `json:"repository"`
	TagPath         string `json:"tag"`
	DigestPath      string `json:"digest"`
	ReleaseNotes    string `json:"releaseNotes"`
	ImagePullSecret string `json:"imagePullSecret"`
}

// Provider - helm3 provider, responsible for managing release updates
type Provider struct {
	implementer Implementer

	sender notification.Sender

	approvalManager approvals.Manager

	events chan *types.Event
	stop   chan struct{}
}

// NewProvider - create new Helm provider
func NewProvider(implementer Implementer, sender notification.Sender, approvalManager approvals.Manager) *Provider {
	return &Provider{
		implementer:     implementer,
		approvalManager: approvalManager,
		sender:          sender,
		events:          make(chan *types.Event, 100),
		stop:            make(chan struct{}),
	}
}

// GetName - get provider name
func (p *Provider) GetName() string {
	return ProviderName
}

// Submit - submit event to provider
func (p *Provider) Submit(event types.Event) error {
	p.events <- &event
	return nil
}

// Start - starts kubernetes provider, waits for events
func (p *Provider) Start() error {
	return p.startInternal()
}

// Stop - stops kubernetes provider
func (p *Provider) Stop() {
	close(p.stop)
}

// TrackedImages - returns tracked images from all releases that have quilla configuration
func (p *Provider) TrackedImages() ([]*types.TrackedImage, error) {
	var trackedImages []*types.TrackedImage

	releases, err := p.implementer.ListReleases()
	if err != nil {
		return nil, err
	}

	for _, release := range releases {
		// getting configuration
		vals, err := values(release.Chart, release.Config)
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err,
				"release":   release.Name,
				"namespace": release.Namespace,
			}).Error("provider.helm3: failed to get values.yaml for release")
			continue
		}

		cfg, err := getquillaConfig(vals)
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err,
				"release":   release.Name,
				"namespace": release.Namespace,
			}).Debug("provider.helm3: failed to get config for release")
			continue
		}

		if cfg.PollSchedule == "" {
			cfg.PollSchedule = types.QuillaPollDefaultSchedule
		}
		// used to check pod secrets
		selector := fmt.Sprintf("app=%s,release=%s", release.Chart.Metadata.Name, release.Name)

		releaseImages, err := getImages(vals)
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err,
				"release":   release.Name,
				"namespace": release.Namespace,
			}).Error("provider.helm3: failed to get images for release")
			continue
		}

		for _, img := range releaseImages {
			img.Meta = map[string]string{
				"selector":      selector,
				"helm.sh/chart": fmt.Sprintf("%s-%s", release.Chart.Metadata.Name, release.Chart.Metadata.Version),
			}
			img.Namespace = release.Namespace
			img.Provider = ProviderName
			trackedImages = append(trackedImages, img)
		}

	}

	return trackedImages, nil
}

func (p *Provider) startInternal() error {
	for {
		select {
		case event := <-p.events:
			err := p.processEvent(event)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"image": event.Repository.Name,
					"tag":   event.Repository.Tag,
				}).Error("provider.helm3: failed to process event")
			}
		case <-p.stop:
			log.Info("provider.helm3: got shutdown signal, stopping...")
			return nil
		}
	}
}

func (p *Provider) processEvent(event *types.Event) (err error) {
	plans, err := p.createUpdatePlans(event)
	if err != nil {
		return err
	}

	approved := p.checkForApprovals(event, plans)

	return p.applyPlans(approved)
}

func (p *Provider) createUpdatePlans(event *types.Event) ([]*UpdatePlan, error) {
	var plans []*UpdatePlan

	releases, err := p.implementer.ListReleases()
	if err != nil {
		return nil, err
	}

	for _, release := range releases {

		plan, update, err := checkRelease(&event.Repository, release.Namespace, release.Name, release.Chart, release.Config)
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err,
				"name":      release.Name,
				"namespace": release.Namespace,
			}).Error("provider.helm3: failed to process versioned release")
			continue
		}

		if update {
			helm3VersionedUpdatesCounter.With(prometheus.Labels{"chart": fmt.Sprintf("%s/%s", release.Namespace, release.Name)}).Inc()
			plans = append(plans, plan)
		}
	}

	return plans, nil
}

func (p *Provider) applyPlans(plans []*UpdatePlan) error {
	for _, plan := range plans {

		p.sender.Send(types.EventNotification{
			ResourceKind: "chart",
			Identifier:   fmt.Sprintf("%s/%s/%s", "chart", plan.Namespace, plan.Name),
			Name:         "update release",
			Message:      fmt.Sprintf("Preparing to update release %s/%s %s->%s (%s)", plan.Namespace, plan.Name, plan.CurrentVersion, plan.NewVersion, strings.Join(mapToSlice(plan.Values), ", ")),
			CreatedAt:    time.Now(),
			Type:         types.NotificationPreReleaseUpdate,
			Level:        types.LevelDebug,
			Channels:     plan.Config.NotificationChannels,
			Metadata: map[string]string{
				"provider":  p.GetName(),
				"namespace": plan.Namespace,
				"name":      plan.Name,
			},
		})

		// err := updateHelmRelease(p.implementer, plan.Name, plan.Chart, plan.Values)
		err := updateHelmRelease(p.implementer, plan.Name, plan.Chart, plan.Values, plan.Namespace, plan.EmptyConfig)
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err,
				"name":      plan.Name,
				"namespace": plan.Namespace,
			}).Error("provider.helm3: failed to apply plan")

			p.sender.Send(types.EventNotification{
				ResourceKind: "chart",
				Identifier:   fmt.Sprintf("%s/%s/%s", "chart", plan.Namespace, plan.Name),
				Name:         "update release",
				Message:      fmt.Sprintf("Release update failed %s/%s %s->%s (%s), error: %s", plan.Namespace, plan.Name, plan.CurrentVersion, plan.NewVersion, strings.Join(mapToSlice(plan.Values), ", "), err),
				CreatedAt:    time.Now(),
				Type:         types.NotificationReleaseUpdate,
				Level:        types.LevelError,
				Channels:     plan.Config.NotificationChannels,
				Metadata: map[string]string{
					"provider":  p.GetName(),
					"namespace": plan.Namespace,
					"name":      plan.Name,
				},
			})
			continue
		}

		err = p.updateComplete(plan)
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err,
				"name":      plan.Name,
				"namespace": plan.Namespace,
			}).Debug("provider.helm3: got error while resetting approvals counter after successful update")
		}

		var msg string
		if len(plan.ReleaseNotes) == 0 {
			msg = fmt.Sprintf("Successfully updated release %s/%s %s->%s (%s)", plan.Namespace, plan.Name, plan.CurrentVersion, plan.NewVersion, strings.Join(mapToSlice(plan.Values), ", "))
		} else {
			msg = fmt.Sprintf("Successfully updated release %s/%s %s->%s (%s). Release notes: %s", plan.Namespace, plan.Name, plan.CurrentVersion, plan.NewVersion, strings.Join(mapToSlice(plan.Values), ", "), strings.Join(plan.ReleaseNotes, ", "))
		}

		p.sender.Send(types.EventNotification{
			ResourceKind: "chart",
			Identifier:   fmt.Sprintf("%s/%s/%s", "chart", plan.Namespace, plan.Name),
			Name:         "update release",
			Message:      msg,
			CreatedAt:    time.Now(),
			Type:         types.NotificationReleaseUpdate,
			Level:        types.LevelSuccess,
			Channels:     plan.Config.NotificationChannels,
			Metadata: map[string]string{
				"provider":  p.GetName(),
				"namespace": plan.Namespace,
				"name":      plan.Name,
			},
		})

	}

	return nil
}

func updateHelmRelease(implementer Implementer, releaseName string, chart *hapi_chart.Chart, overrideValues map[string]string, namespace string, opts ...bool) error {

	// set reuse values to false if currentRelease.config is nil
	emptyConfig := false
	if len(opts) == 1 && opts[0] {
		emptyConfig = opts[0]
	}
	resp, err := implementer.UpdateReleaseFromChart(releaseName, chart, overrideValues, namespace, emptyConfig)

	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"version":        resp.Version,
		"release":        releaseName,
		"overrideValues": overrideValues,
	}).Info("provider.helm3: release updated")
	return nil
}

func mapToSlice(values map[string]string) []string {
	converted := []string{}
	for k, v := range values {
		concat := k + "=" + v
		converted = append(converted, concat)
	}
	return converted
}

// parse
func convertToYaml(values []string) ([]byte, error) {
	base := map[string]interface{}{}
	for _, value := range values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	return yaml.Marshal(base)
}

func getValueAsString(vals chartutil.Values, path string) (string, error) {
	valinterface, err := vals.PathValue(path)
	if err != nil {
		return "", err
	}

	valString := fmt.Sprintf("%v", valinterface)

	return valString, nil
}

// func values(chart *hapi_chart.Chart, config *hapi_chart.Config) (chartutil.Values, error) {
func values(chart *hapi_chart.Chart, config map[string]interface{}) (chartutil.Values, error) {
	return chartutil.CoalesceValues(chart, config)
}

func getquillaConfig(vals chartutil.Values) (*quillaChartConfig, error) {
	yamlFull, err := vals.YAML()
	if err != nil {
		return nil, fmt.Errorf("failed to get vals config, error: %s", err)
	}

	var r Root
	// Default MatchPreRelease to true if not present (backward compatibility)
	r.quilla.MatchPreRelease = true
	err = yaml.Unmarshal([]byte(yamlFull), &r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse quilla config: %s", err)
	}

	if r.quilla.Policy == "" {
		return nil, ErrPolicyNotSpecified
	}

	cfg := r.quilla

	cfg.Plc = policy.GetPolicy(cfg.Policy, &policy.Options{MatchTag: cfg.MatchTag, MatchPreRelease: cfg.MatchPreRelease})

	return &cfg, nil
}
