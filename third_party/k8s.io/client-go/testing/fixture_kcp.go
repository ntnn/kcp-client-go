package testing

import (
	"fmt"
	"sort"

	"github.com/kcp-dev/logicalcluster/v3"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

type ClusterNamespacedName struct {
	Cluster logicalcluster.Path
	types.NamespacedName
}

func (c ClusterNamespacedName) String() string {
	return c.Cluster.String() + "|" + c.NamespacedName.String()
}

// WatchReaction returns a WatchReactionFunc that applies core.Action to
// the given tracker.
func WatchReaction(tracker ObjectTracker) WatchReactionFunc {
	return func(action Action) (bool, watch.Interface, error) {
		cluster := action.GetCluster()
		gvr := action.GetResource()
		ns := action.GetNamespace()
		var watcher watch.Interface
		var err error
		switch cluster {
		case logicalcluster.Wildcard:
			watcher, err = tracker.Watch(gvr, ns)
		default:
			watcher, err = tracker.Cluster(cluster).Watch(gvr, ns)
		}
		if err != nil {
			return false, nil, err
		}
		return true, watcher, nil
	}
}

func (t *tracker) Cluster(clusterPath logicalcluster.Path) ScopedObjectTracker {
	return &scopedTracker{
		tracker:     t,
		clusterPath: clusterPath,
	}
}

type scopedTracker struct {
	*tracker
	clusterPath logicalcluster.Path
}

// AddAll handles adding the objects to the correct place in the tracker, whether they are
// individual items or lists, and handling their logical cluster for you.
func (t *tracker) AddAll(objects ...runtime.Object) error {
	for _, obj := range objects {
		var toAdd []runtime.Object
		if meta.IsListType(obj) {
			list, err := meta.ExtractList(obj)
			if err != nil {
				return err
			}
			errs := runtime.DecodeList(list, t.decoder)
			if len(errs) > 0 {
				return errs[0]
			}
			for _, item := range list {
				toAdd = append(toAdd, item)
			}
		} else {
			toAdd = append(toAdd, obj)
		}

		for i := range toAdd {
			metaObj, ok := toAdd[i].(logicalcluster.Object)
			if !ok {
				return fmt.Errorf("cannot extract logical cluster from %T", toAdd[i])
			}
			if err := t.Cluster(logicalcluster.From(metaObj).Path()).Add(toAdd[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *scopedTracker) List(gvr schema.GroupVersionResource, gvk schema.GroupVersionKind, ns string, opts ...metav1.ListOptions) (runtime.Object, error) {
	list, err := t.List(gvr, gvk, ns, opts...)
	if err != nil {
		return list, err
	}
	objs, err := meta.ExtractList(list)
	if err != nil {
		return nil, err
	}
	matchingObjs, err := filterByCluster(objs, t.clusterPath)
	if err != nil {
		return nil, err
	}
	if err := meta.SetList(list, matchingObjs); err != nil {
		return nil, err
	}
	return list.DeepCopyObject(), nil
}

// filterByCluster returns all objects in the collection that
// match provided namespace. Empty namespace matches
// non-namespaced objects.
func filterByCluster(objs []runtime.Object, cluster logicalcluster.Path) ([]runtime.Object, error) {
	var res []runtime.Object

	for _, obj := range objs {
		acc, err := meta.Accessor(obj)
		if err != nil {
			return nil, err
		}
		if logicalcluster.From(acc).Path() != cluster {
			continue
		}
		res = append(res, obj)
	}

	// Sort res to get deterministic order.
	sort.Slice(res, func(i, j int) bool {
		acc1, _ := meta.Accessor(res[i])
		acc2, _ := meta.Accessor(res[j])
		if acc1.GetNamespace() != acc2.GetNamespace() {
			return acc1.GetNamespace() < acc2.GetNamespace()
		}
		return acc1.GetName() < acc2.GetName()
	})
	return res, nil
}
