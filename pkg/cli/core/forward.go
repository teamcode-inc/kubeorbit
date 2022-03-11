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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kubeorbit.io/pkg/cli/client"
	log "kubeorbit.io/pkg/cli/logger"
	"kubeorbit.io/pkg/cli/util"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func init() {
	setupCloseHandler()
}

var namespace, deploymentName string

func setupCloseHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if namespace != "" && deploymentName != "" {
			log.Infof("uninstall forward workload %s", deploymentName)
			Uninstall(&UninstallRequest{
				Namespace:      namespace,
				DeploymentName: deploymentName,
			})
		}
		os.Exit(0)
	}()
}

type ForwardRequest struct {
	DeploymentName string
	Namespace      string
	LocalPort      int
	ContainerPort  int
}

type Forwarder struct {
	DeploymentName string
	Namespace      string
	LocalPort      int
	ContainerPort  int
	ProxyId        string
	PublicKey      string
	PrivateKey     string
}

func Forward(r *ForwardRequest) error {
	if !util.IsAddrAvailable(fmt.Sprintf(":%d", r.LocalPort)) {
		return fmt.Errorf("local service is not running at %d, please start your service first", r.LocalPort)
	}
	deployment, err := client.KubeClient().AppsV1().Deployments(r.Namespace).Get(context.TODO(), r.DeploymentName, meta.GetOptions{})
	if err != nil {
		return err
	}
	for _, container := range deployment.Spec.Template.Spec.InitContainers {
		if isProxyInitContainer(container.Name) {
			return fmt.Errorf("deployment %s has already forwarded", r.DeploymentName)
		}
	}

	forwarder := r.newForwarder()
	forwarder.wrapDeployment(deployment)

	err = client.KubeClient().AppsV1().Deployments(forwarder.Namespace).Delete(context.TODO(), forwarder.DeploymentName, meta.DeleteOptions{})
	if err != nil {
		return err
	}
	_, err = client.KubeClient().AppsV1().Deployments(forwarder.Namespace).Create(context.TODO(), deployment, meta.CreateOptions{})
	if err != nil {
		return err
	}
	namespace = forwarder.Namespace
	deploymentName = forwarder.DeploymentName
	log.Infof("workload %s recreated", deploymentName)
	defer func() {
		Uninstall(&UninstallRequest{
			Namespace:      namespace,
			DeploymentName: deploymentName,
		})
	}()
	watch, _ := client.KubeClient().CoreV1().Pods(forwarder.Namespace).Watch(context.TODO(), meta.ListOptions{
		LabelSelector: labels.Set(map[string]string{ProxyId: forwarder.ProxyId}).AsSelector().String(),
	})
	for {
		event := <-watch.ResultChan()
		pod, ok := event.Object.(*core.Pod)
		if !ok {
			break
		}

		if pod.Status.Phase == core.PodRunning {
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.Name == ProxyContainer && containerStatus.State.Running != nil {
					channel := &ChannelListener{
						Namespace:  forwarder.Namespace,
						PodName:    pod.Name,
						PrivateKey: forwarder.PrivateKey,
						LocalPort:  forwarder.LocalPort,
					}
					err := channel.ForwardToLocal()
					if err != nil {
						return err
					}
					break
				}
			}
		}
	}
	return nil
}

func (r *ForwardRequest) newForwarder() *Forwarder {
	proxyId := generateProxyId()
	publicKey, privateKey, err := makeSSHKeyPair()
	if err != nil {
		log.Errorf("generate ssh key error")
		return nil
	}
	return &Forwarder{
		DeploymentName: r.DeploymentName,
		Namespace:      r.Namespace,
		LocalPort:      r.LocalPort,
		ContainerPort:  r.ContainerPort,
		ProxyId:        proxyId,
		PublicKey:      publicKey,
		PrivateKey:     privateKey,
	}
}

func (f *Forwarder) wrapDeployment(deployment *apps.Deployment) {
	actualReplicas := deployment.Spec.Replicas
	desireReplicas := getDesiredReplicas()
	initContainers := append(deployment.Spec.Template.Spec.InitContainers, constructInitContainer(f.ContainerPort))
	containers := append(deployment.Spec.Template.Spec.Containers, constructProxyContainer(f.PublicKey))
	deployment.Spec.Replicas = &desireReplicas
	deployment.Spec.Template.Spec.InitContainers = initContainers
	deployment.Spec.Template.Spec.Containers = containers
	deployment.ObjectMeta.Labels[ProxyLabel] = "true"
	deployment.ObjectMeta.Labels[ReplicasLabel] = strconv.Itoa(int(*actualReplicas))
	deployment.ObjectMeta.ResourceVersion = ""
	deployment.Spec.Template.ObjectMeta.Labels[ProxyLabel] = "true"
	deployment.Spec.Template.ObjectMeta.Labels[ProxyId] = f.ProxyId
	deployment.Status = apps.DeploymentStatus{}
}

func makeSSHKeyPair() (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return "", "", err
	}

	// generate and write private key as PEM
	var privateKeyBuf strings.Builder

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(&privateKeyBuf, privateKeyPEM); err != nil {
		return "", "", err
	}

	// generate and write public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	var pubKeyBuf strings.Builder
	pubKeyBuf.Write(ssh.MarshalAuthorizedKey(pub))

	return pubKeyBuf.String(), privateKeyBuf.String(), nil
}
