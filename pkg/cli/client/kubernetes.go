/*
Copyright 2022 The TeamCode authors.
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
package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	log "kubeorbit.io/pkg/cli/logger"
)

var kubeConfig *rest.Config
var kubeClient *kubernetes.Clientset

func init() {
	clusterConfig, err := newClusterConfig()
	kubeConfig = clusterConfig
	if err != nil {
		log.Fatalf("error loading kubeconfig: %v", err)
	}
	clientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalf("error loading kubeconfig: %v", err)
	}
	kubeClient = clientSet
}

func KubeConfig() *rest.Config {
	return kubeConfig
}

func KubeClient() *kubernetes.Clientset {
	return kubeClient
}

func newClusterConfig() (*rest.Config, error) {
	var cfg *rest.Config
	var err error
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	cfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, err
	}

	cfg.QPS = 100
	cfg.Burst = 100

	return cfg, nil
}
