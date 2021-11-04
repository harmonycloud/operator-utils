package util

import (
	"fmt"
	"os"
	"strings"
)

const WATCH_NAMESPACE = "WATCH_NAMESPACE"

func FilterWatchNamespace(namespace string) bool {
	watchNamespace, err := getWatchNamespace()

	if err != nil {
		return false
	}

	if strings.Contains(watchNamespace, ",") {
		nsArray := strings.Split(watchNamespace, ",")
		for _, ns := range nsArray {
			if ns == namespace {
				return true
			}
		}
	} else if namespace == watchNamespace || namespace == "*" {
		return true
	}
	return false
}

// getWatchNamespace returns the Namespace the operator should be watching for changes
func getWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = WATCH_NAMESPACE

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}
