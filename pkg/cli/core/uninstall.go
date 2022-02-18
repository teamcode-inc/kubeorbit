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
	"context"
	apps "k8s.io/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kubeorbit.io/pkg/cli/client"
	log "kubeorbit.io/pkg/cli/logger"
)

type UninstallRequest struct {
	Namespace      string
	DeploymentName string
}

func Uninstall(r *UninstallRequest) error {
	if r.DeploymentName != "" {
		// single deployment
		deployment, err := client.KubeClient().AppsV1().Deployments(r.Namespace).Get(context.TODO(), r.DeploymentName, meta.GetOptions{})
		if err != nil {
			return err
		}
		err = uninstallDeployment(deployment)
		if err != nil {
			return err
		}
	} else {
		// entire namespace
		deployments, err := client.KubeClient().AppsV1().Deployments(r.Namespace).List(context.TODO(), meta.ListOptions{
			LabelSelector: labels.Set(map[string]string{ProxyLabel: "true"}).AsSelector().String(),
		})
		if err != nil {
			return err
		}
		for _, deployment := range deployments.Items {
			err := uninstallDeployment(&deployment)
			if err != nil {
				return err
			}
		}
	}
	log.Infof("workload uninstallation successful")
	return nil
}

func revertDeployment(deployment *apps.Deployment) {
	if deployment.Labels[ProxyLabel] != "true" {
		return
	}
	deployment.Spec.Template.Spec.InitContainers = filterNoneProxyInitContainers(deployment.Spec.Template.Spec.InitContainers)
	deployment.Spec.Template.Spec.Containers = filterNoneProxyContainers(deployment.Spec.Template.Spec.Containers)
	desireReplicas := getDesiredReplicas()
	delete(deployment.ObjectMeta.Labels, ProxyLabel)
	delete(deployment.ObjectMeta.Labels, ReplicasLabel)
	delete(deployment.Spec.Template.ObjectMeta.Labels, ProxyLabel)
	delete(deployment.Spec.Template.ObjectMeta.Labels, ProxyId)
	deployment.ObjectMeta.ResourceVersion = ""
	deployment.Spec.Replicas = &desireReplicas
	deployment.Status = apps.DeploymentStatus{}
}

func uninstallDeployment(deployment *apps.Deployment) error {
	if deployment.Labels[ProxyLabel] != "true" {
		return nil
	}
	revertDeployment(deployment)
	err := client.KubeClient().AppsV1().Deployments(deployment.Namespace).Delete(context.TODO(), deployment.Name, meta.DeleteOptions{})
	if err != nil {
		return err
	}
	_, err = client.KubeClient().AppsV1().Deployments(deployment.Namespace).Create(context.TODO(), deployment, meta.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
