package util

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"strings"
)

const (
	LABEL_MIDD        = "middleware.harmonycloud.cn/app-name"
	LABEL_AUTO_CHANGE = "middleware.harmonycloud.cn/service-autochange"
	// 比如master、slave等
	LABEL_TYPE    = "middleware.harmonycloud.cn/service-type"
	LABEL_STS_POD = "statefulset.kubernetes.io/pod-name"
	LABEL_APP     = "app"

	STATUS_TYPE_KEY    = "type"
	STATUS_TYPE_POD    = "name"
	STATUS_TYPE_STATUS = "status"
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
		podType := condition[STATUS_TYPE_KEY]
		podName := condition[STATUS_TYPE_POD]
		status := condition[STATUS_TYPE_STATUS]
		statusBool := false

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
			//err = UpdateServiceSelector(clientset, namespace, name, "", podType.(string))
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
		map[string]string{LABEL_MIDD: name, LABEL_AUTO_CHANGE: "true"}))

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
				klog.Infof("change %s svc selector to %v", name, selectPodName)
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


func MergeLabel(curLabel,oldLabel map[string]string)map[string]string{
	for k,v := range curLabel{
		oldLabel[k] = v
	}
	return oldLabel
}

func AddAutoChangeLabelsSvc(curLabel map[string]string,name string, svcType string)map[string]string{
	return MergeLabel(curLabel,map[string]string{LABEL_APP: name, LABEL_TYPE: string(svcType)})
}
