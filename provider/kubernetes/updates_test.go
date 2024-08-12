package kubernetes

import (
	"reflect"
	"testing"
	"time"

	"github.com/quilla-hq/quilla/internal/k8s"
	"github.com/quilla-hq/quilla/internal/policy"
	"github.com/quilla-hq/quilla/types"
	"github.com/quilla-hq/quilla/util/timeutil"

	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mustParseGlob(str string) policy.Policy {
	p, err := policy.NewGlobPolicy(str)
	if err != nil {
		panic(err)
	}
	return p
}

func TestProvider_checkForUpdate(t *testing.T) {

	timeutil.Now = func() time.Time {
		return time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
	}
	defer func() { timeutil.Now = time.Now }()

	type args struct {
		policy   policy.Policy
		repo     *types.Repository
		resource *k8s.GenericResource
	}
	tests := []struct {
		name string
		// fields                     fields
		args                       args
		wantUpdatePlan             *UpdatePlan
		wantShouldUpdateDeployment bool
		wantErr                    bool
	}{
		{
			name: "force update untagged to latest",
			args: args{
				policy: policy.NewForcePolicy(false),
				repo:   &types.Repository{Name: "gcr.io/v2-namespace/hello-world", Tag: "latest"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "all"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "all"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world:latest",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
				NewVersion:     "latest",
				CurrentVersion: "latest",
			},
			wantShouldUpdateDeployment: true,
			wantErr:                    false,
		},
		{
			name: "different image name ",
			args: args{
				policy: policy.NewForcePolicy(false),
				repo:   &types.Repository{Name: "gcr.io/v2-namespace/hello-world", Tag: "latest"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "all"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/goodbye-world:earliest",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				// Resource: &k8s.GenericResource{},
				Resource: nil,
			},
			wantShouldUpdateDeployment: false,
			wantErr:                    false,
		},
		{
			name: "different tag name for poll image",
			args: args{
				policy: policy.NewForcePolicy(true),
				repo:   &types.Repository{Name: "gcr.io/v2-namespace/hello-world", Tag: "master"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:      "dep-1",
						Namespace: "xxxx",
						Annotations: map[string]string{
							types.QuillaPollScheduleAnnotation: types.QuillaPollDefaultSchedule,
						},
						Labels: map[string]string{
							types.QuillaPolicyLabel: "all",
						},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world:alpha",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: nil,
			},
			wantShouldUpdateDeployment: false,
			wantErr:                    false,
		},
		{
			name: "dockerhub short image name ",
			args: args{
				policy: policy.NewForcePolicy(false),
				repo:   &types.Repository{Name: "karolisr/quilla", Tag: "0.2.0"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "force"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "karolisr/quilla:latest",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "force"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "karolisr/quilla:0.2.0",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
				NewVersion:     "0.2.0",
				CurrentVersion: "latest",
			},
			wantShouldUpdateDeployment: true,
			wantErr:                    false,
		},
		{
			name: "poll trigger, same tag",
			args: args{
				policy: policy.NewForcePolicy(false),
				repo:   &types.Repository{Name: "karolisr/quilla", Tag: "master"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{types.QuillaPollScheduleAnnotation: types.QuillaPollDefaultSchedule},
						Labels:      map[string]string{types.QuillaPolicyLabel: "force"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "karolisr/quilla:master",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:      "dep-1",
						Namespace: "xxxx",
						Annotations: map[string]string{
							types.QuillaPollScheduleAnnotation: types.QuillaPollDefaultSchedule,
						},
						Labels: map[string]string{types.QuillaPolicyLabel: "force"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "karolisr/quilla:master",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
				NewVersion:     "master",
				CurrentVersion: "master",
			},
			wantShouldUpdateDeployment: true,
			wantErr:                    false,
		},

		{
			name: "pubsub trigger, force-match, same tag",
			args: args{
				policy: policy.NewForcePolicy(false),
				repo:   &types.Repository{Name: "karolisr/quilla", Tag: "latest-staging"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:      "dep-1",
						Namespace: "xxxx",
						Labels: map[string]string{
							types.QuillaPolicyLabel:        "force",
							types.QuillaForceTagMatchLabel: "true",
						},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "karolisr/quilla:latest-staging",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:      "dep-1",
						Namespace: "xxxx",
						Labels: map[string]string{
							types.QuillaForceTagMatchLabel: "true",
							types.QuillaPolicyLabel:        "force",
						},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "karolisr/quilla:latest-staging",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
				NewVersion:     "latest-staging",
				CurrentVersion: "latest-staging",
			},
			wantShouldUpdateDeployment: true,
			wantErr:                    false,
		},

		{
			name: "pubsub trigger, force-match, same tag on eu.gcr.io",
			args: args{
				policy: policy.NewForcePolicy(false),
				repo:   &types.Repository{Host: "eu.gcr.io", Name: "karolisr/quilla", Tag: "latest-staging"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:      "dep-1",
						Namespace: "xxxx",
						Labels: map[string]string{
							types.QuillaPolicyLabel:        "force",
							types.QuillaForceTagMatchLabel: "true",
						},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "eu.gcr.io/karolisr/quilla:latest-staging",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:      "dep-1",
						Namespace: "xxxx",
						Labels: map[string]string{
							types.QuillaForceTagMatchLabel: "true",
							types.QuillaPolicyLabel:        "force",
						},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
									// "time": timeutil.Now().String(),
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "eu.gcr.io/karolisr/quilla:latest-staging",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
				NewVersion:     "latest-staging",
				CurrentVersion: "latest-staging",
			},
			wantShouldUpdateDeployment: true,
			wantErr:                    false,
		},
		{
			name: "pubsub trigger, force-match, different tag",
			args: args{
				policy: policy.NewForcePolicy(true),
				repo:   &types.Repository{Name: "karolisr/quilla", Tag: "latest-staging"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:      "dep-1",
						Namespace: "xxxx",
						Labels: map[string]string{
							types.QuillaPolicyLabel:        "force",
							types.QuillaForceTagMatchLabel: "true",
						},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "karolisr/quilla:latest-acceptance",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: nil,
			},
			wantShouldUpdateDeployment: false,
			wantErr:                    false,
		},
		{
			name: "pubsub trigger, force-match, same tag on eu.gcr.io, daemonset",
			args: args{
				policy: policy.NewForcePolicy(false),
				repo:   &types.Repository{Host: "eu.gcr.io", Name: "karolisr/quilla", Tag: "latest-staging"},
				resource: MustParseGR(&apps_v1.DaemonSet{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:      "dep-1",
						Namespace: "xxxx",
						Labels:    map[string]string{types.QuillaPolicyLabel: "force"},
						Annotations: map[string]string{
							types.QuillaForceTagMatchLabel: "true",
						},
					},
					apps_v1.DaemonSetSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "eu.gcr.io/karolisr/quilla:latest-staging",
									},
								},
							},
						},
					},
					apps_v1.DaemonSetStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: MustParseGR(&apps_v1.DaemonSet{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:      "dep-1",
						Namespace: "xxxx",
						Annotations: map[string]string{
							types.QuillaForceTagMatchLabel: "true",
						},
						Labels: map[string]string{types.QuillaPolicyLabel: "force"},
					},
					apps_v1.DaemonSetSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
									// "time": timeutil.Now().String(),
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "eu.gcr.io/karolisr/quilla:latest-staging",
									},
								},
							},
						},
					},
					apps_v1.DaemonSetStatus{},
				}),
				NewVersion:     "latest-staging",
				CurrentVersion: "latest-staging",
			},
			wantShouldUpdateDeployment: true,
			wantErr:                    false,
		},
		{
			name: "daemonset, glob matcher",
			args: args{
				policy: mustParseGlob("glob:release-*"),
				repo:   &types.Repository{Host: "eu.gcr.io", Name: "karolisr/quilla", Tag: "release-2"},
				resource: MustParseGR(&apps_v1.DaemonSet{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:      "dep-1",
						Namespace: "xxxx",
						Labels:    map[string]string{types.QuillaPolicyLabel: "glob:release-*"},
						Annotations: map[string]string{
							types.QuillaForceTagMatchLabel: "true",
						},
					},
					apps_v1.DaemonSetSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "eu.gcr.io/karolisr/quilla:release-1",
									},
								},
							},
						},
					},
					apps_v1.DaemonSetStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: MustParseGR(&apps_v1.DaemonSet{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:      "dep-1",
						Namespace: "xxxx",
						Annotations: map[string]string{
							types.QuillaForceTagMatchLabel: "true",
						},
						Labels: map[string]string{types.QuillaPolicyLabel: "glob:release-*"},
					},
					apps_v1.DaemonSetSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
									// "time": timeutil.Now().String(),
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "eu.gcr.io/karolisr/quilla:release-2",
									},
								},
							},
						},
					},
					apps_v1.DaemonSetStatus{},
				}),
				NewVersion:     "release-2",
				CurrentVersion: "release-1",
			},
			wantShouldUpdateDeployment: true,
			wantErr:                    false,
		},

		{
			name: "update init container if tracking is enabled",
			args: args{
				policy: policy.NewForcePolicy(false),
				repo:   &types.Repository{Name: "gcr.io/v2-namespace/hello-world", Tag: "latest"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{types.QuillaInitContainerAnnotation: "true"},
						Labels:      map[string]string{types.QuillaPolicyLabel: "all"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								InitContainers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{types.QuillaInitContainerAnnotation: "true"},
						Labels:      map[string]string{types.QuillaPolicyLabel: "all"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								InitContainers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world:latest",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
				NewVersion:     "latest",
				CurrentVersion: "latest",
			},
			wantShouldUpdateDeployment: true,
			wantErr:                    false,
		},
		{
			name: "do not update init container if tracking is disabled (default)",
			args: args{
				policy: policy.NewForcePolicy(false),
				repo:   &types.Repository{Name: "gcr.io/v2-namespace/hello-world", Tag: "latest"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "all"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								InitContainers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				// Resource: &k8s.GenericResource{},
				Resource: nil,
			},
			wantShouldUpdateDeployment: false,
			wantErr:                    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUpdatePlan, gotShouldUpdateDeployment, err := checkForUpdate(tt.args.policy, tt.args.repo, tt.args.resource)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.checkUnversionedDeployment() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}

			if gotShouldUpdateDeployment {
				ann := gotUpdatePlan.Resource.GetSpecAnnotations()

				if ann[types.QuillaUpdateTimeAnnotation] != "" {
					delete(ann, types.QuillaUpdateTimeAnnotation)
					gotUpdatePlan.Resource.SetSpecAnnotations(ann)
				} else {
					t.Errorf("Provider.checkUnversionedDeployment() missing types.QuillaUpdateTimeAnnotation annotation")
				}
			}

			if !reflect.DeepEqual(gotUpdatePlan, tt.wantUpdatePlan) {
				t.Errorf("Provider.checkUnversionedDeployment() gotUpdatePlan = %#v, want %#v", gotUpdatePlan, tt.wantUpdatePlan)
			}
			if gotShouldUpdateDeployment != tt.wantShouldUpdateDeployment {
				t.Errorf("Provider.checkUnversionedDeployment() gotShouldUpdateDeployment = %#v, want %#v", gotShouldUpdateDeployment, tt.wantShouldUpdateDeployment)
			}
		})
	}
}

func TestProvider_checkForUpdateSemver(t *testing.T) {

	type args struct {
		policy   policy.Policy
		repo     *types.Repository
		resource *k8s.GenericResource
	}
	tests := []struct {
		name                       string
		args                       args
		wantUpdatePlan             *UpdatePlan
		wantShouldUpdateDeployment bool
		wantErr                    bool
	}{
		{
			name: "standard version bump",
			args: args{
				policy: policy.NewSemverPolicy(policy.SemverPolicyTypeAll, true),
				repo:   &types.Repository{Name: "gcr.io/v2-namespace/hello-world", Tag: "1.1.2"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "all"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world:1.1.1",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "all"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world:1.1.2",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
				NewVersion:     "1.1.2",
				CurrentVersion: "1.1.1",
			},
			wantShouldUpdateDeployment: true,
			wantErr:                    false,
		},
		{
			name: "staging pre-release",
			args: args{

				policy: policy.NewSemverPolicy(policy.SemverPolicyTypeMinor, true),
				repo:   &types.Repository{Name: "gcr.io/v2-namespace/hello-prerelease", Tag: "v1.1.2-staging"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "minor"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-prerelease:v1.1.1",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan:             &UpdatePlan{},
			wantShouldUpdateDeployment: false,
			wantErr:                    false,
		},
		{
			name: "normal new tag while there's pre-release",
			args: args{

				policy: policy.NewSemverPolicy(policy.SemverPolicyTypeMinor, true),
				repo:   &types.Repository{Name: "gcr.io/v2-namespace/hello-prerelease", Tag: "v1.1.2"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "minor"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-prerelease:v1.1.1-staging",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan:             &UpdatePlan{},
			wantShouldUpdateDeployment: false,
			wantErr:                    false,
		},
		{
			name: "standard ignore version bump",
			args: args{

				policy: policy.NewSemverPolicy(policy.SemverPolicyTypeAll, true),
				repo:   &types.Repository{Name: "gcr.io/v2-namespace/hello-world", Tag: "1.1.1"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "all"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world:1.1.1",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource:       nil,
				NewVersion:     "",
				CurrentVersion: "",
			},
			wantShouldUpdateDeployment: false,
			wantErr:                    false,
		},
		{
			name: "multiple containers, version bump one",
			args: args{
				policy: policy.NewSemverPolicy(policy.SemverPolicyTypeAll, true),
				repo:   &types.Repository{Name: "gcr.io/v2-namespace/hello-world", Tag: "1.1.2"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "all"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world:1.1.1",
									},
									{
										Image: "yo-world:1.1.1",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "all"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world:1.1.2",
									},
									{
										Image: "yo-world:1.1.1",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
				NewVersion:     "1.1.2",
				CurrentVersion: "1.1.1",
			},
			wantShouldUpdateDeployment: true,
			wantErr:                    false,
		},
		{
			name: "force update untagged container",
			args: args{
				policy: policy.NewForcePolicy(false),
				repo:   &types.Repository{Name: "gcr.io/v2-namespace/hello-world", Tag: "1.1.2"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "force"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world:latest",
									},
									{
										Image: "yo-world:1.1.1",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels:      map[string]string{types.QuillaPolicyLabel: "force"},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world:1.1.2",
									},
									{
										Image: "yo-world:1.1.1",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
				NewVersion:     "1.1.2",
				CurrentVersion: "latest",
			},
			wantShouldUpdateDeployment: true,
			wantErr:                    false,
		},
		{
			name: "force update untagged container - match tag",
			args: args{
				policy: policy.NewForcePolicy(true),
				repo:   &types.Repository{Name: "gcr.io/v2-namespace/hello-world", Tag: "1.1.2"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels: map[string]string{
							types.QuillaPolicyLabel:        "force",
							types.QuillaForceTagMatchLabel: "true",
						},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world:1.1.2",
									},
									{
										Image: "yo-world:1.1.1",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels: map[string]string{
							types.QuillaPolicyLabel:        "force",
							types.QuillaForceTagMatchLabel: "true",
						},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world:1.1.2",
									},
									{
										Image: "yo-world:1.1.1",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
				NewVersion:     "1.1.2",
				CurrentVersion: "1.1.2",
			},
			wantShouldUpdateDeployment: true,
			wantErr:                    false,
		},
		{
			name: "don't force update untagged container - match tag",
			args: args{
				policy: policy.NewForcePolicy(true),
				repo:   &types.Repository{Name: "gcr.io/v2-namespace/hello-world", Tag: "1.1.3"},
				resource: MustParseGR(&apps_v1.Deployment{
					meta_v1.TypeMeta{},
					meta_v1.ObjectMeta{
						Name:        "dep-1",
						Namespace:   "xxxx",
						Annotations: map[string]string{},
						Labels: map[string]string{
							types.QuillaPolicyLabel:        "force",
							types.QuillaForceTagMatchLabel: "true",
						},
					},
					apps_v1.DeploymentSpec{
						Template: v1.PodTemplateSpec{
							ObjectMeta: meta_v1.ObjectMeta{
								Annotations: map[string]string{
									"this": "that",
								},
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "gcr.io/v2-namespace/hello-world:1.1.2",
									},
									{
										Image: "yo-world:1.1.1",
									},
								},
							},
						},
					},
					apps_v1.DeploymentStatus{},
				}),
			},
			wantUpdatePlan: &UpdatePlan{
				Resource:       nil,
				NewVersion:     "",
				CurrentVersion: "",
			},
			wantShouldUpdateDeployment: false,
			wantErr:                    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUpdatePlan, gotShouldUpdateDeployment, err := checkForUpdate(tt.args.policy, tt.args.repo, tt.args.resource)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.checkVersionedDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotShouldUpdateDeployment {
				ann := gotUpdatePlan.Resource.GetSpecAnnotations()
				_, ok := ann[types.QuillaUpdateTimeAnnotation]
				if ok {
					delete(ann, types.QuillaUpdateTimeAnnotation)
					gotUpdatePlan.Resource.SetSpecAnnotations(ann)
				} else {
					t.Errorf("Provider.checkVersionedDeployment() missing types.QuillaUpdateTimeAnnotation annotation")
				}
			}

			if !reflect.DeepEqual(gotUpdatePlan, tt.wantUpdatePlan) {
				t.Errorf("Provider.checkVersionedDeployment() gotUpdatePlan = %v, want %v", gotUpdatePlan, tt.wantUpdatePlan)
			}
			if gotShouldUpdateDeployment != tt.wantShouldUpdateDeployment {
				t.Errorf("Provider.checkVersionedDeployment() gotShouldUpdateDeployment = %v, want %v", gotShouldUpdateDeployment, tt.wantShouldUpdateDeployment)
			}
		})
	}
}
