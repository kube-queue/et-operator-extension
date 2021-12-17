/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/kube-queue/et-operator-extension/pkg/et-operator/apis/et/v1alpha1"
	scheme "github.com/kube-queue/et-operator-extension/pkg/et-operator/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// TrainingJobsGetter has a method to return a TrainingJobInterface.
// A group's client should implement this interface.
type TrainingJobsGetter interface {
	TrainingJobs(namespace string) TrainingJobInterface
}

// TrainingJobInterface has methods to work with TrainingJob resources.
type TrainingJobInterface interface {
	Create(ctx context.Context, trainingJob *v1alpha1.TrainingJob, opts v1.CreateOptions) (*v1alpha1.TrainingJob, error)
	Update(ctx context.Context, trainingJob *v1alpha1.TrainingJob, opts v1.UpdateOptions) (*v1alpha1.TrainingJob, error)
	UpdateStatus(ctx context.Context, trainingJob *v1alpha1.TrainingJob, opts v1.UpdateOptions) (*v1alpha1.TrainingJob, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.TrainingJob, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.TrainingJobList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.TrainingJob, err error)
	TrainingJobExpansion
}

// trainingJobs implements TrainingJobInterface
type trainingJobs struct {
	client rest.Interface
	ns     string
}

// newTrainingJobs returns a TrainingJobs
func newTrainingJobs(c *EtV1alpha1Client, namespace string) *trainingJobs {
	return &trainingJobs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the trainingJob, and returns the corresponding trainingJob object, and an error if there is any.
func (c *trainingJobs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.TrainingJob, err error) {
	result = &v1alpha1.TrainingJob{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of TrainingJobs that match those selectors.
func (c *trainingJobs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.TrainingJobList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.TrainingJobList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("trainingjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested trainingJobs.
func (c *trainingJobs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("trainingjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a trainingJob and creates it.  Returns the server's representation of the trainingJob, and an error, if there is any.
func (c *trainingJobs) Create(ctx context.Context, trainingJob *v1alpha1.TrainingJob, opts v1.CreateOptions) (result *v1alpha1.TrainingJob, err error) {
	result = &v1alpha1.TrainingJob{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("trainingjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(trainingJob).
		Do().
		Into(result)
	return
}

// Update takes the representation of a trainingJob and updates it. Returns the server's representation of the trainingJob, and an error, if there is any.
func (c *trainingJobs) Update(ctx context.Context, trainingJob *v1alpha1.TrainingJob, opts v1.UpdateOptions) (result *v1alpha1.TrainingJob, err error) {
	result = &v1alpha1.TrainingJob{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(trainingJob.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(trainingJob).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *trainingJobs) UpdateStatus(ctx context.Context, trainingJob *v1alpha1.TrainingJob, opts v1.UpdateOptions) (result *v1alpha1.TrainingJob, err error) {
	result = &v1alpha1.TrainingJob{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(trainingJob.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(trainingJob).
		Do().
		Into(result)
	return
}

// Delete takes name of the trainingJob and deletes it. Returns an error if one occurs.
func (c *trainingJobs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(name).
		Body(&opts).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *trainingJobs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("trainingjobs").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do().
		Error()
}

// Patch applies the patch and returns the patched trainingJob.
func (c *trainingJobs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.TrainingJob, err error) {
	result = &v1alpha1.TrainingJob{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do().
		Into(result)
	return
}
