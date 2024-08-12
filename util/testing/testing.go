package testing

import (
	"fmt"

	"github.com/quilla-hq/quilla/internal/k8s"
	"github.com/quilla-hq/quilla/types"
	"github.com/quilla-hq/quilla/util/image"

	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// FakeK8sImplementer - fake implementer used for testing
type FakeK8sImplementer struct {
	NamespacesList   *v1.NamespaceList
	DeploymentSingle *apps_v1.Deployment
	DeploymentList   *apps_v1.DeploymentList

	// stores value of an updated deployment
	Updated *k8s.GenericResource

	AvailableSecret map[string]*v1.Secret

	AvailablePods *v1.PodList
	DeletedPods   []*v1.Pod

	// error to return
	Error error
}

// Namespaces - available namespaces
func (i *FakeK8sImplementer) Namespaces() (*v1.NamespaceList, error) {
	return i.NamespacesList, nil
}

// Deployment - available deployment, doesn't filter anything
func (i *FakeK8sImplementer) Deployment(namespace, name string) (*apps_v1.Deployment, error) {
	return i.DeploymentSingle, nil
}

// Deployments - available deployments
func (i *FakeK8sImplementer) Deployments(namespace string) (*apps_v1.DeploymentList, error) {
	return i.DeploymentList, nil
}

// Update - update deployment
func (i *FakeK8sImplementer) Update(obj *k8s.GenericResource) error {
	i.Updated = obj
	return nil
}

// Secret - get secret
func (i *FakeK8sImplementer) Secret(namespace, name string) (*v1.Secret, error) {
	if i.Error != nil {
		return nil, i.Error
	}
	s, ok := i.AvailableSecret[name]
	if !ok {
		return nil, fmt.Errorf("secret %s not found", name)
	}
	return s, nil
}

// Pods - available pods
func (i *FakeK8sImplementer) Pods(namespace, labelSelector string) (*v1.PodList, error) {
	return i.AvailablePods, nil
}

// ConfigMaps - returns nothing (not implemented)
func (i *FakeK8sImplementer) ConfigMaps(namespace string) core_v1.ConfigMapInterface {
	panic("not implemented")
}

// DeletePod - adds pod to DeletedPods list
func (i *FakeK8sImplementer) DeletePod(namespace, name string, opts *meta_v1.DeleteOptions) error {
	i.DeletedPods = append(i.DeletedPods, &v1.Pod{
		meta_v1.TypeMeta{},
		meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		v1.PodSpec{},
		v1.PodStatus{},
	})
	return nil
}

func GetTrackedImage(i string) *types.TrackedImage {
	ref, err := image.Parse(i)
	if err != nil {
		panic(err)
	}
	return &types.TrackedImage{
		Image:        ref,
		PollSchedule: "",
		Trigger:      types.TriggerTypeDefault,
		Provider:     "",
		Namespace:    "",
		Meta:         make(map[string]string),
		Tags:         []string{ref.Tag()},
	}
}
