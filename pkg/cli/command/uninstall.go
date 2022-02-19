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
package command

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"kubeorbit.io/pkg/cli/client"
	"kubeorbit.io/pkg/cli/core"
)

func UninstallCommand() *cobra.Command {
	clientCfg, _ := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	namespace := clientCfg.Contexts[clientCfg.CurrentContext].Namespace

	if namespace == "" {
		namespace = "default"
	}
	request := &core.UninstallRequest{}
	cmd := &cobra.Command{
		Use:  "uninstall",
		Long: `Uninstall orbit agent and resources`,
		Run: func(cmd *cobra.Command, args []string) {
			err := core.Uninstall(request)
			if err != nil {
				cmd.PrintErr(err)
			}
		},
	}
	cmd.Flags().StringVar(&request.Namespace, "namespace", client.GetDefaultNamespace(), "Namespace for uninstall")
	cmd.Flags().StringVar(&request.DeploymentName, "deployment", "", "Deployment name for uninstall")
	return cmd
}
