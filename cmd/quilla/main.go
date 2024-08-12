package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"context"

	"github.com/prometheus/client_golang/prometheus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/quilla-hq/quilla/approvals"
	"github.com/quilla-hq/quilla/bot"

	// "github.com/quilla-hq/quilla/cache/memory"
	"github.com/quilla-hq/quilla/pkg/auth"
	"github.com/quilla-hq/quilla/pkg/http"
	"github.com/quilla-hq/quilla/pkg/store"
	"github.com/quilla-hq/quilla/pkg/store/sql"

	"github.com/quilla-hq/quilla/constants"
	"github.com/quilla-hq/quilla/extension/credentialshelper"
	"github.com/quilla-hq/quilla/extension/notification"
	"github.com/quilla-hq/quilla/internal/k8s"
	"github.com/quilla-hq/quilla/internal/workgroup"
	"github.com/quilla-hq/quilla/provider"
	"github.com/quilla-hq/quilla/provider/helm3"
	"github.com/quilla-hq/quilla/provider/kubernetes"
	"github.com/quilla-hq/quilla/registry"
	"github.com/quilla-hq/quilla/secrets"
	"github.com/quilla-hq/quilla/trigger/poll"
	"github.com/quilla-hq/quilla/trigger/pubsub"
	"github.com/quilla-hq/quilla/types"
	"github.com/quilla-hq/quilla/version"

	// notification extensions
	"github.com/quilla-hq/quilla/extension/notification/auditor"
	_ "github.com/quilla-hq/quilla/extension/notification/discord"
	_ "github.com/quilla-hq/quilla/extension/notification/hipchat"
	_ "github.com/quilla-hq/quilla/extension/notification/mail"
	_ "github.com/quilla-hq/quilla/extension/notification/mattermost"
	_ "github.com/quilla-hq/quilla/extension/notification/slack"
	_ "github.com/quilla-hq/quilla/extension/notification/teams"
	_ "github.com/quilla-hq/quilla/extension/notification/webhook"

	// credentials helpers
	_ "github.com/quilla-hq/quilla/extension/credentialshelper/aws"
	_ "github.com/quilla-hq/quilla/extension/credentialshelper/azure"
	_ "github.com/quilla-hq/quilla/extension/credentialshelper/gcr"
	secretsCredentialsHelper "github.com/quilla-hq/quilla/extension/credentialshelper/secrets"

	// bots
	_ "github.com/quilla-hq/quilla/bot/hipchat"
	_ "github.com/quilla-hq/quilla/bot/slack"

	log "github.com/sirupsen/logrus"
	// importing to ensure correct dependencies
	_ "helm.sh/helm/v3/pkg/action"
)

// gcloud pubsub related config
const (
	EnvTriggerPubSub = "PUBSUB" // set to 1 or something to enable pub/sub trigger
	EnvTriggerPoll   = "POLL"   // set to 0 to disable poll trigger
	EnvProjectID     = "PROJECT_ID"
	EnvClusterName   = "CLUSTER_NAME"
	EnvDataDir       = "XDG_DATA_HOME"
	EnvHelm3Provider = "HELM3_PROVIDER" // helm3 provider
	EnvUIDir         = "UI_DIR"

	// EnvDefaultDockerRegistryCfg - default registry configuration that can be passed into
	// quilla for polling trigger
	EnvDefaultDockerRegistryCfg = "DOCKER_REGISTRY_CFG"
)

// kubernetes config, if empty - will default to InCluster
const (
	EnvKubernetesConfig = "KUBERNETES_CONFIG"
)

// EnvDebug - set to 1 or anything else to enable debug logging
const EnvDebug = "DEBUG"

func main() {
	ver := version.GetquillaVersion()

	inCluster := kingpin.Flag("incluster", "use in cluster configuration (defaults to 'true'), use '--no-incluster' if running outside of the cluster").Default("true").Bool()
	kubeconfig := kingpin.Flag("kubeconfig", "path to kubeconfig (if not in running inside a cluster)").Default(filepath.Join(os.Getenv("HOME"), ".kube", "config")).String()
	uiDir := kingpin.Flag("ui-dir", "path to web UI static files").Default("www").Envar(EnvUIDir).String()

	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version(ver.Version)
	kingpin.CommandLine.Help = "Automated Kubernetes deployment updates. Learn more on https://quilla.sh."
	kingpin.Parse()

	log.WithFields(log.Fields{
		"os":         ver.OS,
		"build_date": ver.BuildDate,
		"revision":   ver.Revision,
		"version":    ver.Version,
		"go_version": ver.GoVersion,
		"arch":       ver.Arch,
	}).Info("quilla starting...")

	if os.Getenv(EnvDebug) == "true" {
		log.SetLevel(log.DebugLevel)
	}

	dataDir := "/data"
	if os.Getenv(EnvDataDir) != "" {
		dataDir = os.Getenv(EnvDataDir)
	}

	sqlStore, err := sql.New(sql.Opts{
		DatabaseType: "sqlite3",
		URI:          filepath.Join(dataDir, "quilla.db"),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to initialize database")
		os.Exit(1)
	}
	log.WithFields(log.Fields{
		"database_path": filepath.Join(dataDir, "quilla.db"),
		"type":          "sqlite3",
	}).Info("initializing database")

	// registering auditor to log events
	auditLogger := auditor.New(sqlStore)
	notification.RegisterSender("auditor", auditLogger)

	// setting up triggers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notificationLevel := types.LevelInfo
	if os.Getenv(constants.EnvNotificationLevel) != "" {
		parsedLevel, err := types.ParseLevel(os.Getenv(constants.EnvNotificationLevel))
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Errorf("main: got error while parsing notification level, defaulting to: %s", notificationLevel)
		} else {
			notificationLevel = parsedLevel
		}
	}

	notifCfg := &notification.Config{
		Attempts: 10,
		Level:    notificationLevel,
	}
	sender := notification.New(ctx)

	_, err = sender.Configure(notifCfg)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("main: failed to configure notification sender manager")
	}

	// getting k8s provider
	k8sCfg := &kubernetes.Opts{
		ConfigPath: *kubeconfig,
	}

	if os.Getenv(EnvKubernetesConfig) != "" {
		k8sCfg.ConfigPath = os.Getenv(EnvKubernetesConfig)
	}

	k8sCfg.InCluster = *inCluster

	implementer, err := kubernetes.NewKubernetesImplementer(k8sCfg)
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"config": k8sCfg,
		}).Fatal("main: failed to create kubernetes implementer")
	}

	var g workgroup.Group

	t := &k8s.Translator{
		FieldLogger: log.WithField("context", "translator"),
	}

	buf := k8s.NewBuffer(&g, t, log.StandardLogger(), 128)
	wl := log.WithField("context", "watch")
	k8s.WatchDeployments(&g, implementer.Client(), wl, buf)
	k8s.WatchStatefulSets(&g, implementer.Client(), wl, buf)
	k8s.WatchDaemonSets(&g, implementer.Client(), wl, buf)
	k8s.WatchCronJobs(&g, implementer.Client(), wl, buf)

	// approvalsCache := memory.NewMemoryCache()
	approvalsManager := approvals.New(&approvals.Opts{
		// Cache: approvalsCache,
		Store: sqlStore,
	})

	pendindApprovalsCounter := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "pending_approvals",
		Help: "Number of the pending approvals",
	}, func() float64 {
		approvals, err := approvalsManager.List()
		if err != nil {
			return float64(len(approvals))
		}
		return 0
	})
	prometheus.MustRegister(pendindApprovalsCounter)

	go approvalsManager.StartExpiryService(ctx)

	// setting up providers
	providers := setupProviders(&ProviderOpts{
		k8sImplementer:   implementer,
		sender:           sender,
		approvalsManager: approvalsManager,
		grc:              &t.GenericResourceCache,
		store:            sqlStore,
		k8sClient:        implementer.Client(),
		config:           implementer.Config(),
	})

	// registering secrets based credentials helper
	dockerConfig := make(secrets.DockerCfg)
	if os.Getenv(EnvDefaultDockerRegistryCfg) != "" {
		dockerConfigStr := os.Getenv(EnvDefaultDockerRegistryCfg)
		log.Infof("docker config: %s: ", dockerConfigStr)
		dockerConfig, err = secrets.DecodeDockerCfgJson([]byte(dockerConfigStr))
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatalf("failed to decode secret provided in %s env variable", EnvDefaultDockerRegistryCfg)
		}
	}
	secretsGetter := secrets.NewGetter(implementer, dockerConfig)

	ch := secretsCredentialsHelper.New(secretsGetter)
	credentialshelper.RegisterCredentialsHelper("secrets", ch)

	// trigger setup
	// teardownTriggers := setupTriggers(ctx, providers, approvalsManager, &t.GenericResourceCache, implementer)
	teardownTriggers := setupTriggers(ctx, &TriggerOpts{
		providers:        providers,
		approvalsManager: approvalsManager,
		grc:              &t.GenericResourceCache,
		k8sClient:        implementer,
		store:            sqlStore,
		uiDir:            *uiDir,
	})

	bot.Run(implementer, approvalsManager)

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	g.Add(func(stop <-chan struct{}) {
		go func() {
			for range signalChan {
				log.Info("received an interrupt, shutting down...")
				go func() {
					select {
					case <-time.After(10 * time.Second):
						log.Info("connection shutdown took too long, exiting... ")
						close(cleanupDone)
						return
					case <-cleanupDone:
						return
					}
				}()
				providers.Stop()
				teardownTriggers()
				bot.Stop()

				cleanupDone <- true
			}
		}()
		<-cleanupDone
	})
	g.Run()
}

type ProviderOpts struct {
	k8sImplementer   kubernetes.Implementer
	sender           notification.Sender
	approvalsManager approvals.Manager
	grc              *k8s.GenericResourceCache
	store            store.Store

	k8sClient kube.Interface
	config    *rest.Config
}

// setupProviders - setting up available providers. New providers should be initialised here and added to
// provider map
func setupProviders(opts *ProviderOpts) (providers provider.Providers) {
	var enabledProviders []provider.Provider

	k8sProvider, err := kubernetes.NewProvider(opts.k8sImplementer, opts.sender, opts.approvalsManager, opts.grc)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("main.setupProviders: failed to create kubernetes provider")
	}
	go func() {
		err := k8sProvider.Start()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("kubernetes provider stopped with an error")
		}
	}()

	enabledProviders = append(enabledProviders, k8sProvider)

	if os.Getenv(EnvHelm3Provider) == "1" || os.Getenv(EnvHelm3Provider) == "true" {
		helm3Implementer := helm3.NewHelm3Implementer()
		helm3Provider := helm3.NewProvider(helm3Implementer, opts.sender, opts.approvalsManager)

		go func() {
			err := helm3Provider.Start()
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Fatal("helm3 provider stopped with an error")
			}
		}()

		enabledProviders = append(enabledProviders, helm3Provider)

	}

	providers = provider.New(enabledProviders, opts.approvalsManager)

	return providers
}

type TriggerOpts struct {
	providers        provider.Providers
	approvalsManager approvals.Manager
	grc              *k8s.GenericResourceCache
	k8sClient        kubernetes.Implementer
	store            store.Store
	uiDir            string
}

// setupTriggers - setting up triggers. New triggers should be added to this function. Each trigger
// should go through all providers (or not if there is a reason) and submit events)
// func setupTriggers(ctx context.Context, providers provider.Providers, approvalsManager approvals.Manager, grc *k8s.GenericResourceCache, k8sClient kubernetes.Implementer) (teardown func()) {
func setupTriggers(ctx context.Context, opts *TriggerOpts) (teardown func()) {
	issuers := map[string]auth.Issuer{}
	enabled, err := strconv.ParseBool(os.Getenv(constants.EnvAzureAADEnabled))
	if err == nil && enabled {
		issuers[fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", os.Getenv(constants.EnvAzureTenantID))] = auth.Issuer{
			Jwks:          fmt.Sprintf("https://login.microsoftonline.com/%s/discovery/v2.0/keys", os.Getenv(constants.EnvAzureTenantID)),
			Name:          "Azure",
			UsernameClaim: "preferred_username",
		}
	}

	if os.Getenv(constants.EnvBasicAuthUser) != "" && os.Getenv(constants.EnvBasicAuthPassword) != "" {
		issuers["quilla"] = auth.Issuer{
			Jwks:          "",
			Name:          "Quilla",
			UsernameClaim: "username",
		}
	}

	var authenticator auth.Authenticator
	authenticator = auth.New(&auth.Opts{
		Username: os.Getenv(constants.EnvBasicAuthUser),
		Password: os.Getenv(constants.EnvBasicAuthPassword),
		Secret:   []byte(os.Getenv(constants.EnvTokenSecret)),
	}, issuers)

	enabled, err = strconv.ParseBool(os.Getenv(constants.EnvRBACEnabled))
	// setting up generic http webhook server
	whs := http.NewTriggerServer(&http.Opts{
		Port:                  types.QuillaDefaultPort,
		GRC:                   opts.grc,
		KubernetesClient:      opts.k8sClient,
		Providers:             opts.providers,
		ApprovalManager:       opts.approvalsManager,
		Store:                 opts.store,
		Authenticator:         authenticator,
		UIDir:                 opts.uiDir,
		AuthenticatedWebhooks: os.Getenv(constants.EnvAuthenticatedWebhooks) == "true",
		RBACEnabled:           err == nil && enabled,
	})

	go func() {
		err := whs.Start()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"port":  types.QuillaDefaultPort,
			}).Fatal("trigger server stopped")
		}
	}()

	// checking whether pubsub (GCR) trigger is enabled
	if os.Getenv(EnvTriggerPubSub) != "" {
		projectID := os.Getenv(EnvProjectID)
		if projectID == "" {
			log.Fatalf("main.setupTriggers: project ID env variable not set")
			return
		}

		ps, err := pubsub.NewPubsubSubscriber(&pubsub.Opts{
			ProjectID: projectID,
			Providers: opts.providers,
		})
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("main.setupTriggers: failed to create gcloud pubsub subscriber")
			return
		}

		subManager := pubsub.NewDefaultManager(os.Getenv(EnvClusterName), projectID, opts.providers, ps)
		go subManager.Start(ctx)
	}

	if os.Getenv(EnvTriggerPoll) != "0" || os.Getenv(EnvTriggerPoll) != "false" {

		registryClient := registry.New()
		watcher := poll.NewRepositoryWatcher(opts.providers, registryClient)
		pollManager := poll.NewPollManager(opts.providers, watcher)

		// start poll manager, will finish with ctx
		go watcher.Start(ctx)
		go pollManager.Start(ctx)
	}

	teardown = func() {
		whs.Stop()
	}

	return teardown
}
