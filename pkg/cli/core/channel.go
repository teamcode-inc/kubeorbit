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
	"golang.org/x/crypto/ssh"
	"io"
	"k8s.io/apimachinery/pkg/util/rand"
	log "kubeorbit.io/pkg/cli/logger"
	"kubeorbit.io/pkg/cli/util"
	"net"
	"sync"
	"time"
)

func connectSSHServer(address, privateKey string) (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, err
	}
	client, err := ssh.Dial("tcp", address, &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func findAvailablePort() (int, error) {
	for i := 0; i < 10; i++ {
		port := rand.Intn(65535-1024) + 1024
		if !util.IsAddrAvailable(fmt.Sprintf(":%d", port)) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("cannot find available port")
}

type ChannelListener struct {
	Namespace  string
	PodName    string
	PrivateKey string
	LocalPort  int
}

func (c *ChannelListener) ForwardToLocal() error {
	stop := make(chan struct{})
	sshForwardPort, err := findAvailablePort()
	go func() {
		err = portForward(c.Namespace, c.PodName, sshForwardPort, ProxySSHPort, stop)
		if err != nil {
			log.Errorf("port forward err: %v", err)
		}
	}()
	sshForwardAddress := fmt.Sprintf(":%d", sshForwardPort)
	localPortAddress := fmt.Sprintf(":%d", c.LocalPort)
	containerProxyAddress := fmt.Sprintf("0.0.0.0:%d", ProxyPort)
	err = waitForAddress(sshForwardAddress, 30*time.Second)
	if err != nil {
		return err
	}
	log.Infof("ssh forwarded localPort %d to remotePort %d", sshForwardPort, ProxySSHPort)
	sshClient, err := connectSSHServer(sshForwardAddress, c.PrivateKey)
	if err != nil {
		return err
	}
	listener, err := sshClient.Listen("tcp", containerProxyAddress)
	if err != nil {
		return err
	}
	log.Infof("channel connected, you can start testing your service")
	for {
		sshConn, err := listener.Accept()
		if err != nil {
			log.Errorf("connection err %v", err)
			break
		}
		localConn, err := net.Dial("tcp", localPortAddress)
		if err != nil {
			sshConn.Close()
			break
			log.Errorf("connection err %v", err)
		}
		go func() {
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						wg.Done()
					}
				}()
				io.Copy(sshConn, localConn)
				wg.Done()
				sshConn.Close()
			}()
			wg.Add(1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						wg.Done()
					}
				}()
				io.Copy(localConn, sshConn)
				wg.Done()
				localConn.Close()
			}()
			wg.Wait()
		}()
	}
	return nil
}

func waitForAddress(address string, timeOut time.Duration) error {
	var depChan = make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		go func(address string) {
			defer wg.Done()
			for {
				conn, err := net.Dial("tcp", address)
				if err == nil {
					conn.Close()
					return
				}
				time.Sleep(1 * time.Second)
			}
		}(address)
		wg.Wait()
		close(depChan)
	}()

	select {
	case <-depChan:
		return nil
	case <-time.After(timeOut):
		return fmt.Errorf("address %s aren't ready in %d", address, timeOut)
	}
}
