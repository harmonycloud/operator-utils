package resource

import (
	"context"
	"encoding/json"
	apps "k8s.io/api/apps/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	kutil "github.com/harmonycloud/operator-utils/pkg/util"
)

func CreateOrPatchStatefulSet(ctx context.Context, c kubernetes.Interface, meta metav1.ObjectMeta,transform func(*apps.StatefulSet) *apps.StatefulSet, opts metav1.PatchOptions) (*apps.StatefulSet, kutil.VerbType, error) {
	cur, err := c.AppsV1().StatefulSets(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		klog.V(3).Infof("Creating StatefulSet %s/%s.", meta.Namespace, meta.Name)
		out, err := c.AppsV1().StatefulSets(meta.Namespace).Create(ctx, transform(&apps.StatefulSet{
			TypeMeta: metav1.TypeMeta{
				Kind:       "StatefulSet",
				APIVersion: apps.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}), metav1.CreateOptions{
			DryRun:       opts.DryRun,
			FieldManager: opts.FieldManager,
		})
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchStatefulSet(ctx, c, cur, transform, opts)
}


func PatchStatefulSet(ctx context.Context, c kubernetes.Interface, cur *apps.StatefulSet, transform func(*apps.StatefulSet) *apps.StatefulSet, opts metav1.PatchOptions) (*apps.StatefulSet, kutil.VerbType, error) {
	return PatchStatefulSetObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
}

func PatchStatefulSetObject(ctx context.Context, c kubernetes.Interface, cur, mod *apps.StatefulSet, opts metav1.PatchOptions) (*apps.StatefulSet, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, apps.StatefulSet{})
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	klog.V(3).Infof("Patching StatefulSet %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.AppsV1().StatefulSets(cur.Namespace).Patch(ctx, cur.Name, types.StrategicMergePatchType, patch, opts)
	return out, kutil.VerbPatched, err
}