package fake

import (
	"context"

	"github.com/kcp-dev/logicalcluster/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/metadata"

	kcptesting "github.com/kcp-dev/client-go/third_party/k8s.io/client-go/testing"
)

var _ kcpmetadata.ClusterInterface = (*FakeMetadataClusterClientset)(nil)
var _ kcptesting.FakeClient = (*FakeMetadataClusterClientset)(nil)

// FakeMetadataClusterClientset implements clientset.Interface. Meant to be embedded into a
// struct to get a default implementation. This makes faking out just the method
// you want to test easier.
type FakeMetadataClusterClientset struct {
	*kcptesting.Fake
	scheme  *runtime.Scheme
	tracker kcptesting.ObjectTracker
}

func (c *FakeMetadataClusterClientset) Tracker() kcptesting.ObjectTracker {
	return c.tracker
}

func (c *FakeMetadataClusterClientset) Cluster(clusterPath logicalcluster.Path) metadata.Interface {
	if clusterPath == logicalcluster.Wildcard {
		panic("A specific cluster must be provided when scoping, not the wildcard.")
	}
	return c.cluster(clusterPath)
}

func (c *FakeMetadataClusterClientset) cluster(clusterPath logicalcluster.Path) metadata.Interface {
	return &FakeMetadataClient{
		Fake:        c.Fake,
		tracker:     c.tracker.Cluster(clusterPath),
		clusterPath: clusterPath,
	}
}

func (c *FakeMetadataClusterClientset) Resource(resource schema.GroupVersionResource) kcpmetadata.ResourceClusterInterface {
	return &FakeMetadataClusterClient{
		Fake:     c.Fake,
		scheme:   c.scheme,
		tracker:  c.tracker,
		resource: resource,
	}
}

type FakeMetadataClusterClient struct {
	*kcptesting.Fake
	scheme   *runtime.Scheme
	tracker  kcptesting.ObjectTracker
	resource schema.GroupVersionResource
}

func (f *FakeMetadataClusterClient) Cluster(clusterPath logicalcluster.Path) metadata.Getter {
	if clusterPath == logicalcluster.Wildcard {
		panic("A specific cluster must be provided when scoping, not the wildcard.")
	}
	return f.cluster(clusterPath)
}

func (f *FakeMetadataClusterClient) cluster(clusterPath logicalcluster.Path) metadata.Getter {
	return &metadataResourceClient{
		client: &FakeMetadataClient{
			Fake:        f.Fake,
			scheme:      f.scheme,
			tracker:     f.tracker.Cluster(clusterPath),
			clusterPath: clusterPath,
		},
		resource: f.resource,
	}
}

func (f *FakeMetadataClusterClient) List(ctx context.Context, opts metav1.ListOptions) (*metav1.PartialObjectMetadataList, error) {
	return f.cluster(logicalcluster.Wildcard).List(ctx, opts)
}

func (f *FakeMetadataClusterClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return f.cluster(logicalcluster.Wildcard).Watch(ctx, opts)
}
