package fake

import (
	"context"

	"github.com/kcp-dev/logicalcluster/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"

	kcpdynamic "github.com/kcp-dev/client-go/dynamic"
	kcptesting "github.com/kcp-dev/client-go/third_party/k8s.io/client-go/testing"
)

// TODO add boilerplate

var (
	_ kcpdynamic.ClusterInterface = &FakeDynamicClusterClientset{}
	_ kcptesting.FakeClient       = &FakeDynamicClusterClientset{}
)

type FakeDynamicClusterClientset struct {
	*kcptesting.Fake
	scheme        *runtime.Scheme
	gvrToListKind map[schema.GroupVersionResource]string
	tracker       kcptesting.ObjectTracker
}

func (c *FakeDynamicClusterClientset) Tracker() kcptesting.ObjectTracker {
	return c.tracker
}

func (c *FakeDynamicClusterClientset) Cluster(clusterPath logicalcluster.Path) dynamic.Interface {
	if clusterPath == logicalcluster.Wildcard {
		panic("A specific cluster must be provided when scoping, not the wildcard.")
	}
	return &FakeDynamicClient{
		Fake:          c.Fake,
		tracker:       c.tracker.Cluster(clusterPath),
		clusterPath:   clusterPath,
		gvrToListKind: c.gvrToListKind,
	}
}

func (c *FakeDynamicClusterClientset) Resource(resource schema.GroupVersionResource) kcpdynamic.ResourceClusterInterface {
	return &FakeDynamicClusterClient{
		Fake:          c.Fake,
		scheme:        c.scheme,
		gvrToListKind: c.gvrToListKind,
		tracker:       c.tracker,
		resource:      resource,
	}
}

var (
	_ kcpdynamic.ResourceClusterInterface = &FakeDynamicClusterClient{}
	_ kcptesting.FakeClient               = &FakeDynamicClusterClient{}
)

type FakeDynamicClusterClient struct {
	*kcptesting.Fake
	scheme        *runtime.Scheme
	gvrToListKind map[schema.GroupVersionResource]string
	tracker       kcptesting.ObjectTracker
	resource      schema.GroupVersionResource
}

func (f *FakeDynamicClusterClient) Tracker() kcptesting.ObjectTracker {
	return f.tracker
}

func (f *FakeDynamicClusterClient) Cluster(clusterPath logicalcluster.Path) dynamic.NamespaceableResourceInterface {
	if clusterPath == logicalcluster.Wildcard {
		panic("A specific cluster must be provided when scoping, not the wildcard.")
	}
	return f.cluster(clusterPath)
}

func (f *FakeDynamicClusterClient) cluster(clusterPath logicalcluster.Path) dynamic.NamespaceableResourceInterface {
	return &dynamicResourceClient{
		client: &FakeDynamicClient{
			Fake:          f.Fake,
			tracker:       f.tracker.Cluster(clusterPath),
			clusterPath:   clusterPath,
			gvrToListKind: f.gvrToListKind,
		},
		resource: f.resource,
		listKind: f.gvrToListKind[f.resource],
	}
}

func (f *FakeDynamicClusterClient) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return f.cluster(logicalcluster.Wildcard).List(ctx, opts)
}

func (f *FakeDynamicClusterClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return f.cluster(logicalcluster.Wildcard).Watch(ctx, opts)
}
