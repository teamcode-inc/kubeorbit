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
package core

import (
	"fmt"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"kubeorbit.io/pkg/cli/client"
	"net/http"
)

func portForward(namespace, podName string, localPort, remotePort int, stop chan struct{}) error {
	req := client.KubeClient().CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(client.KubeConfig())
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, req.URL())
	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}
	ready := make(chan struct{})
	fw, err := portforward.New(dialer, ports, stop, ready, nil, nil)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}
