package maintain

import (
	"context"
	"github.com/harmonycloud/operator-utils/pkg/util"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const (
	LABEL_STS_POD = "statefulset.kubernetes.io/pod-name"
	LABEL_APP     = "app"

	STATUS_TYPE_KEY    = "type"
	STATUS_TYPE_POD    = "name"
	STATUS_TYPE_STATUS = "status"
)

var (
	// 比如master、slave等
	LABEL_TYPE    = util.AdaptCompany("middleware.harmonycloud.cn/service-type")
)

/**
 * @Author shenkonghui
 * @Description 直接利用status.condition进行svc选择
 * @Date 11:05 2021/9/14
 * @Param
 * @return
 **/
func UpdateServiceSelectorWithArray(conditions []map[string]interface{},
	clientset clientset.Interface,
	namespace string,
	name string) error {

	for _, condition := range conditions {
		podType, exist1 := condition[STATUS_TYPE_KEY]
		podName, exist2 := condition[STATUS_TYPE_POD]
		status, _ := condition[STATUS_TYPE_STATUS]
		statusBool := false
		if !exist1 || !exist2  {
			klog.Info("condition format err")
			return nil
		}

		// 为了兼容，status可能为字符串也有可能为bool类型
		if status != nil {
			switch status.(type) {
			case string:
				if status.(string) == "true" {
					statusBool = true
				}
			case bool:
				if status.(bool) == true {
					statusBool = true
				}
			}
		}

		var err error
		if statusBool == true {
			err = UpdateServiceSelector(clientset, namespace, name, podName.(string), podType.(string))
		} else {
			// 状态异常，直接设置pod为空，不选任何pod
			err = UpdateServiceSelector(clientset, namespace, name, "", podType.(string))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

/**
 * @Author shenkonghui
 * @Description 动态修改service的selector中的statefulset.kubernetes.io/pod-name，选择特定的pod
 * @Date 11:05 2021/9/14
 * @Param
 * @return
 **/
func UpdateServiceSelector(clientset clientset.Interface,
	namespace string,
	name string,
	selectPodName string,
	svcType string) error {

	var labelSelector labels.Selector
	labelSelector = labels.SelectorFromSet(labels.Set(
		map[string]string{LABEL_APP: name}))

	services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector.String()})
	if err != nil {
		return err
	}

	for _, service := range services.Items {
		isMatch := false
		if svcType != "" {
			serviceType := service.Labels[LABEL_TYPE]
			// labels不支持"/"等特殊字符，只能采用"_"
			serviceTypes := strings.Split(serviceType, "_")
			for _, t := range serviceTypes {
				// 匹配到类型
				if svcType == t {
					isMatch = true
					break
				}
			}
		}

		if isMatch {
			if service.Spec.Selector[LABEL_STS_POD] != selectPodName {
				klog.Infof("change [%s] svc selector to [%v]", service.Name, selectPodName)
				service.Spec.Selector[LABEL_STS_POD] = selectPodName
				_, err := clientset.CoreV1().Services(namespace).Update(context.TODO(), &service, metav1.UpdateOptions{})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func MergeLabel(curLabel, oldLabel map[string]string) map[string]string {
	for k, v := range curLabel {
		oldLabel[k] = v
	}
	return oldLabel
}

func AddAutoChangeLabelsSvc(curLabel map[string]string, name string, svcType string) map[string]string {
	return MergeLabel(curLabel, map[string]string{LABEL_APP: name, LABEL_TYPE: string(svcType)})
}

func VerifyService(svc *v1.Service) bool {
	if svc == nil {
		return false
	}

	_, exist := svc.Labels[LABEL_APP]
	if !exist {
		return false
	}
	_, exist = svc.Labels[LABEL_TYPE]
	if !exist {
		return false
	}
	return true
}
