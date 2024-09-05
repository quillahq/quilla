package kubernetes

import (
	"context"
	"fmt"
	"os"

	"github.com/quilla-hq/quilla/internal/k8s"

	apps_v1 "k8s.io/api/apps/v1"
	batch_v1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	log "github.com/sirupsen/logrus"
)

// Implementer - thing wrapper around currently used k8s APIs
type Implementer interface {
	Namespaces() (*v1.NamespaceList, error)
	Deployments(namespace string) (*apps_v1.DeploymentList, error)
	Update(obj *k8s.GenericResource) error
	Secret(namespace, name string) (*v1.Secret, error)
	Pods(namespace, labelSelector string) (*v1.PodList, error)
	DeletePod(namespace, name string, opts *meta_v1.DeleteOptions) error
	CreateJob(name string, image string, secret string) error
	Job(namespace, name string) (*batch_v1.Job, error)

	ConfigMaps(namespace string) core_v1.ConfigMapInterface
}

// KubernetesImplementer - default kubernetes client implementer, uses
// https://github.com/kubernetes/client-go v3.0.0-beta.0
type KubernetesImplementer struct {
	cfg    *rest.Config
	client *kubernetes.Clientset
}

// Opts - implementer options, usually for k8s deployments
// it's best to use InCluster option
type Opts struct {
	// if set - kube config options will be ignored
	InCluster  bool
	ConfigPath string
	Master     string
}

// NewKubernetesImplementer - create new k8s implementer
func NewKubernetesImplementer(opts *Opts) (*KubernetesImplementer, error) {
	cfg := &rest.Config{}

	if opts.InCluster {
		var err error
		cfg, err = rest.InClusterConfig()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("provider.kubernetes: failed to get kubernetes config")
			return nil, err
		}
		log.Info("provider.kubernetes: using in-cluster configuration")
	} else if opts.ConfigPath != "" {
		var err error
		cfg, err = clientcmd.BuildConfigFromFlags("", opts.ConfigPath)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("provider.kubernetes: failed to get cmd kubernetes config")
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("kubernetes config is missing")
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("provider.kubernetes: failed to create kubernetes client")
		return nil, err
	}

	return &KubernetesImplementer{client: client, cfg: cfg}, nil
}

func (i *KubernetesImplementer) Client() *kubernetes.Clientset {
	return i.client
}

func (i *KubernetesImplementer) Config() *rest.Config {
	return i.cfg
}

// Namespaces - get all namespaces
func (i *KubernetesImplementer) Namespaces() (*v1.NamespaceList, error) {
	namespaces := i.client.CoreV1().Namespaces()
	return namespaces.List(context.TODO(), meta_v1.ListOptions{})
}

// Deployment - get specific deployment for namespace/name
func (i *KubernetesImplementer) Deployment(namespace, name string) (*apps_v1.Deployment, error) {
	dep := i.client.AppsV1().Deployments(namespace)
	return dep.Get(context.TODO(), name, meta_v1.GetOptions{})
}

// Deployments - get all deployments for namespace
func (i *KubernetesImplementer) Deployments(namespace string) (*apps_v1.DeploymentList, error) {
	dep := i.client.AppsV1().Deployments(namespace)
	l, err := dep.List(context.TODO(), meta_v1.ListOptions{})
	return l, err
}

// Update converts generic resource into specific kubernetes type and updates it
func (i *KubernetesImplementer) Update(obj *k8s.GenericResource) error {
	// retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
	// 	// Retrieve the latest version of Deployment before attempting update
	// 	// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
	// 	_, updateErr := i.client.Extensions().Deployments(deployment.Namespace).Update(deployment)
	// 	return updateErr
	// })
	// return retryErr

	switch resource := obj.GetResource().(type) {
	case *apps_v1.Deployment:
		_, err := i.client.AppsV1().Deployments(resource.Namespace).Update(context.TODO(), resource, meta_v1.UpdateOptions{})
		if err != nil {
			return err
		}
	case *apps_v1.StatefulSet:
		_, err := i.client.AppsV1().StatefulSets(resource.Namespace).Update(context.TODO(), resource, meta_v1.UpdateOptions{})
		if err != nil {
			return err
		}
	case *apps_v1.DaemonSet:
		_, err := i.client.AppsV1().DaemonSets(resource.Namespace).Update(context.TODO(), resource, meta_v1.UpdateOptions{})
		if err != nil {
			return err
		}
	case *batch_v1.CronJob:
		_, err := i.client.BatchV1().CronJobs(resource.Namespace).Update(context.TODO(), resource, meta_v1.UpdateOptions{})
		if err != nil {
			return err
		}
	case *batch_v1.Job:
		_, err := i.client.BatchV1().Jobs(resource.Namespace).Update(context.TODO(), resource, meta_v1.UpdateOptions{})
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported object type")
	}
	return nil
}

func (i *KubernetesImplementer) CreateJob(name string, image string, secret string) error {
	ns := os.Getenv("NAMESPACE")
	if ns == "" {
		ns = "default"
	}

	var backoff int32 = 1

	jobSpec := &batch_v1.Job{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      fmt.Sprintf("gate-job-%s", name),
			Namespace: ns,
		},
		Spec: batch_v1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "job",
							Image:           image,
							ImagePullPolicy: v1.PullIfNotPresent,
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
			BackoffLimit: &backoff,
		},
	}

	if secret != "" {
		jobSpec.Spec.Template.Spec.Containers[0].EnvFrom = []v1.EnvFromSource{
			{
				SecretRef: &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: secret,
					},
				},
			},
		}
	}

	_, err := i.client.BatchV1().Jobs(ns).Create(context.TODO(), jobSpec, meta_v1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (i *KubernetesImplementer) Job(namespace, name string) (*batch_v1.Job, error) {
	return i.client.BatchV1().Jobs(namespace).Get(context.TODO(), fmt.Sprintf("gate-job-%s", name), meta_v1.GetOptions{})
}

// Secret - get secret
func (i *KubernetesImplementer) Secret(namespace, name string) (*v1.Secret, error) {
	return i.client.CoreV1().Secrets(namespace).Get(context.TODO(), name, meta_v1.GetOptions{})
}

// Pods - get pods
func (i *KubernetesImplementer) Pods(namespace, labelSelector string) (*v1.PodList, error) {
	return i.client.CoreV1().Pods(namespace).List(context.TODO(), meta_v1.ListOptions{LabelSelector: labelSelector})
}

// DeletePod - delete pod by name
func (i *KubernetesImplementer) DeletePod(namespace, name string, opts *meta_v1.DeleteOptions) error {
	return i.client.CoreV1().Pods(namespace).Delete(context.TODO(), name, *opts)
}

// ConfigMaps - returns an interface to config maps for a specified namespace
func (i *KubernetesImplementer) ConfigMaps(namespace string) core_v1.ConfigMapInterface {
	return i.client.CoreV1().ConfigMaps(namespace)
}
