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
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	uuid "github.com/satori/go.uuid"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
)

const (
	ProxyInitContainer = "kubeorbit-proxy-init"
	ProxyContainer     = "kubeorbit-proxy"
	ProxyLabel         = "kubeorbit.io/proxy-forward"
	ProxyId            = "kubeorbit.io/proxy-id"
	ReplicasLabel      = "kubeorbit.io/workload-actual-replicas"
	ProxySSHPort       = 2222
	ProxyPort          = 18201
)

func constructInitContainer(containerPort int) core.Container {
	privileged := true
	runAsUser := int64(0)
	runAsGroup := int64(0)
	return core.Container{
		Name:            ProxyInitContainer,
		Image:           "soarinferret/iptablesproxy:latest",
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         []string{"iptables"},
		Args: []string{
			"-t",
			"nat",
			"-A",
			"PREROUTING",
			"-p",
			"tcp",
			"--dport",
			//TODO forward all port by default
			strconv.Itoa(containerPort),
			"-j",
			"REDIRECT",
			"--to-ports",
			strconv.Itoa(ProxyPort),
		},
		Resources: core.ResourceRequirements{
			Limits: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("100m"),
				core.ResourceMemory: resource.MustParse("50Mi"),
			},
			Requests: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("10m"),
				core.ResourceMemory: resource.MustParse("10Mi"),
			},
		},
		SecurityContext: &core.SecurityContext{
			Capabilities: &core.Capabilities{
				Add: []core.Capability{
					"NET_ADMIN",
					"NET_RAW",
				},
				Drop: []core.Capability{
					"ALL",
				},
			},
			Privileged: &privileged,
			RunAsUser:  &runAsUser,
			RunAsGroup: &runAsGroup,
		},
	}
}

func constructProxyContainer(RSAPublicKey string) core.Container {
	return core.Container{
		Name:            ProxyContainer,
		Image:           "teamcode2021/orbit-proxy:latest",
		ImagePullPolicy: core.PullIfNotPresent,
		Resources: core.ResourceRequirements{
			Limits: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("100m"),
				core.ResourceMemory: resource.MustParse("50Mi"),
			},
			Requests: core.ResourceList{
				core.ResourceCPU:    resource.MustParse("10m"),
				core.ResourceMemory: resource.MustParse("10Mi"),
			},
		},
		Env: []core.EnvVar{
			{
				Name:  "RSAPublicKey",
				Value: base64.StdEncoding.EncodeToString([]byte(RSAPublicKey)),
			},
		},
		Ports: []core.ContainerPort{
			{
				ContainerPort: ProxySSHPort,
			},
			{
				ContainerPort: ProxyPort,
			},
		},
		LivenessProbe: &core.Probe{
			ProbeHandler: core.ProbeHandler{
				TCPSocket: &core.TCPSocketAction{
					Port: intstr.IntOrString{
						IntVal: ProxySSHPort,
					},
				},
			},
		},
	}
}

func isProxyInitContainer(containerName string) bool {
	if containerName == ProxyInitContainer {
		return true
	}
	return false
}

func filterNoneProxyInitContainers(containers []core.Container) []core.Container {
	var initContainers []core.Container
	for _, container := range containers {
		if container.Name != ProxyInitContainer {
			initContainers = append(initContainers, container)
		}
	}
	return initContainers
}

func filterNoneProxyContainers(containers []core.Container) []core.Container {
	var initContainers []core.Container
	for _, container := range containers {
		if container.Name != ProxyContainer {
			initContainers = append(initContainers, container)
		}
	}
	return initContainers
}

func getDesiredReplicas() int32 {
	return 1
}

func generateProxyId() string {
	uuid := uuid.NewV4()
	hash := md5.Sum(uuid.Bytes())
	return hex.EncodeToString(hash[:])[8:16]
}
