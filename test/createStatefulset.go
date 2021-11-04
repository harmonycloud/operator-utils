package test

import (
	"context"
	"github.com/harmonycloud/operator-utils/pkg/resource"
	"github.com/harmonycloud/operator-utils/pkg/util/merge"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)


var Client kubernetes.Interface

func main(){
	Mongo := MongoDBCommunity{
		Spec: MongoDBCommunitySpec{
		},
	}
	statefulSetMeta := metav1.ObjectMeta{
		Name:      Mongo.Name,
		Namespace: Mongo.Namespace,
	}

	// 构建statefulset
	sts,_,_:= resource.CreateOrPatchStatefulSet(context.TODO(),Client,statefulSetMeta, buildStatefulset, metav1.PatchOptions{})
	// 扩展自定义的stateulset属性
	sts.Spec = merge.StatefulSetSpecs(sts.Spec, Mongo.Spec.StatefulSetConfiguration.SpecWrapper.Spec)
}

func buildStatefulset (in *appsv1.StatefulSet) *appsv1.StatefulSet {
	return in
}