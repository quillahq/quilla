// Package types holds most of the types used across quilla
//
//go:generate jsonenums -type=Notification
//go:generate jsonenums -type=Level
//go:generate jsonenums -type=TriggerType
//go:generate jsonenums -type=ProviderType
package types

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// quillaDefaultPort - default port for application
const QuillaDefaultPort = 9300

// quillaPolicyLabel - quilla update policies (version checking)
const QuillaPolicyLabel = "quilla.sh/policy"

const QuillaImagePullSecretAnnotation = "quilla.sh/imagePullSecret"

// quillaTriggerLabel - trigger label is used to specify custom trigger types
// for example quilla.sh/trigger=poll would signal poll trigger to start watching for repository
// changes
const QuillaTriggerLabel = "quilla.sh/trigger"

// quillaForceTagMatchLabel - label that checks whether tags match before force updating
const QuillaForceTagMatchLegacyLabel = "quilla.sh/match-tag"
const QuillaForceTagMatchLabel = "quilla.sh/matchTag"

// quillaMatchPreReleaseAnnotation - label or annotation to set pre-release matching for SemVer, defaults to true for backward compatibility
const QuillaMatchPreReleaseAnnotation = "quilla.sh/matchPreRelease"

// quillaPollScheduleAnnotation - optional variable to setup custom schedule for polling, defaults to @every 10m
const QuillaPollScheduleAnnotation = "quilla.sh/pollSchedule"

// quillaInitContainerAnnotation - label or annotation to track init containers, defaults to false for backward compatibility
const QuillaInitContainerAnnotation = "quilla.sh/initContainers"

// quillaPollDefaultSchedule - defaul polling schedule
var QuillaPollDefaultSchedule = "@every 1m"

// quillaDigestAnnotation - digest annotation
const QuillaDigestAnnotation = "quilla.sh/digest"

// quillaNotificationChanAnnotation - optional notification to override
// default notification channel(-s) per deployment/chart
const QuillaNotificationChanAnnotation = "quilla.sh/notify"

// quillaMinimumApprovalsLabel - min approvals
const QuillaMinimumApprovalsLabel = "quilla.sh/approvals"

// quillaUpdateTimeAnnotation - update time
const QuillaUpdateTimeAnnotation = "quilla.sh/update-time"

// quillaApprovalDeadlineLabel - approval deadline
const QuillaApprovalDeadlineLabel = "quilla.sh/approvalDeadline"

// quillaApprovalDeadlineDefault - default deadline in hours
const QuillaApprovalDeadlineDefault = 24

// quillaReleasePage - optional release notes URL passed on with notification
const QuillaReleaseNotesURL = "quilla.sh/releaseNotes"

func init() {
	value, found := os.LookupEnv("POLL_DEFAULTSCHEDULE")
	if found {
		QuillaPollDefaultSchedule = value
	}
}

// Repository - represents main docker repository fields that
// quilla cares about
type Repository struct {
	Host   string `json:"host"`
	Name   string `json:"name"`
	Tag    string `json:"tag"`
	Digest string `json:"digest"` // optional digest field
}

// String gives you [host/]team/repo[:tag] identifier
func (r *Repository) String() string {
	b := bytes.NewBufferString(r.Host)
	if b.Len() != 0 {
		b.WriteRune('/')
	}
	b.WriteString(r.Name)
	if r.Tag != "" {
		b.WriteRune(':')
		b.WriteString(r.Tag)
	}
	return b.String()
}

// Event - holds information about new event from trigger
type Event struct {
	Repository Repository `json:"repository,omitempty"`
	CreatedAt  time.Time  `json:"createdAt,omitempty"`
	// optional field to identify trigger
	TriggerName string `json:"triggerName,omitempty"`
}

func (e *Event) Value() (driver.Value, error) {
	j, err := json.Marshal(e)
	return j, err
}

func (e *Event) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed.")
	}

	var event Event
	if err := json.Unmarshal(source, &event); err != nil {
		return err
	}

	*e = event

	return nil
}

// Version - version container
type Version struct {
	Major      int64
	Minor      int64
	Patch      int64
	PreRelease string
	Metadata   string

	Original string
}

func (v Version) String() string {
	if v.Original != "" {
		return v.Original
	}
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		fmt.Fprintf(&buf, "-%s", v.PreRelease)
	}
	if v.Metadata != "" {
		fmt.Fprintf(&buf, "+%s", v.Metadata)
	}

	return buf.String()

}

// TriggerType - trigger types
type TriggerType int

// Available trigger types
const (
	TriggerTypeDefault  TriggerType = iota // default policy is to wait for external triggers
	TriggerTypePoll                        // poll policy sets up watchers for the affected repositories
	TriggerTypeApproval                    // fulfilled approval requests trigger events
)

func (t TriggerType) String() string {
	switch t {
	case TriggerTypeDefault:
		return "default"
	case TriggerTypePoll:
		return "poll"
	case TriggerTypeApproval:
		return "approval"
	default:
		return "default"
	}
}

// ParseTrigger - parse trigger string into type
func ParseTrigger(trigger string) TriggerType {
	switch trigger {
	case "poll":
		return TriggerTypePoll
	}
	return TriggerTypeDefault
}

// EventNotification notification used for sending
type EventNotification struct {
	Name         string       `json:"name"`
	Message      string       `json:"message"`
	CreatedAt    time.Time    `json:"createdAt"`
	Type         Notification `json:"type"`
	Level        Level        `json:"level"`
	ResourceKind string       `json:"resourceKind"`
	Identifier   string       `json:"identifier"`
	// Channels is an optional variable to override
	// default channel(-s) when performing an update
	Channels []string `json:"-"`

	Metadata map[string]string `json:"metadata"`
}

// ParseEventNotificationChannels - parses deployment annotations  or chart config
// to get channel overrides
func ParseEventNotificationChannels(annotations map[string]string) []string {
	channels := []string{}
	if annotations == nil {
		return channels
	}
	chanStr, ok := annotations[QuillaNotificationChanAnnotation]
	if ok {
		chans := strings.Split(chanStr, ",")
		for _, c := range chans {
			channels = append(channels, strings.TrimSpace(c))
		}
	}

	return channels
}

func ParseReleaseNotesURL(annotations map[string]string) string {
	if annotations == nil {
		return ""
	}

	return annotations[QuillaReleaseNotesURL]
}

// Notification - notification types used by notifier
type Notification int

// available notification types for hooks
const (
	PreProviderSubmitNotification Notification = iota
	PostProviderSubmitNotification

	// Kubernetes notification types
	NotificationPreDeploymentUpdate
	NotificationDeploymentUpdate

	// Helm notification types
	NotificationPreReleaseUpdate
	NotificationReleaseUpdate

	NotificationSystemEvent

	NotificationUpdateApproved
	NotificationUpdateRejected
)

func (n Notification) String() string {
	switch n {
	case PreProviderSubmitNotification:
		return "pre provider submit"
	case PostProviderSubmitNotification:
		return "post provider submit"
	case NotificationPreDeploymentUpdate:
		return "preparing deployment update"
	case NotificationDeploymentUpdate:
		return "deployment update"
	case NotificationPreReleaseUpdate:
		return "preparing release update"
	case NotificationReleaseUpdate:
		return "release update"
	case NotificationSystemEvent:
		return "system event"
	case NotificationUpdateApproved:
		return "update approved"
	case NotificationUpdateRejected:
		return "update rejected "
	default:
		return "unknown"
	}
}

// Level - event levet
type Level int

// Available event levels
const (
	LevelDebug Level = iota
	LevelInfo
	LevelSuccess
	LevelWarn
	LevelError
	LevelFatal
)

// ParseLevel takes a string level and returns notification level constant.
func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "fatal":
		return LevelFatal, nil
	case "error":
		return LevelError, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "info":
		return LevelInfo, nil
	case "success":
		return LevelSuccess, nil
	case "debug":
		return LevelDebug, nil
	}

	var l Level
	return l, fmt.Errorf("not a valid notification Level: %q", lvl)
}

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelSuccess:
		return "success"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// Color - used to assign different colors for events
func (l Level) Color() string {
	switch l {
	case LevelError:
		return "#F44336"
	case LevelInfo:
		return "#2196F3"
	case LevelSuccess:
		return "#00C853"
	case LevelFatal:
		return "#B71C1C"
	case LevelWarn:
		return "#FF9800"
	default:
		return "#9E9E9E"
	}
}

// ProviderType - provider type used to differentiate different providers
// when used with plugins
type ProviderType int

// Known provider types
const (
	ProviderTypeUnknown ProviderType = iota
	ProviderTypeKubernetes
	ProviderTypeHelm
)

func (t ProviderType) String() string {
	switch t {
	case ProviderTypeUnknown:
		return "unknown"
	case ProviderTypeKubernetes:
		return "kubernetes"
	case ProviderTypeHelm:
		return "helm"
	default:
		return ""
	}
}
