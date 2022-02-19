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
	"kubeorbit.io/pkg/cli/client"
	"kubeorbit.io/pkg/cli/core"
)

func ForwardCommand() *cobra.Command {

	request := &core.ForwardRequest{}
	cmd := &cobra.Command{
		Use:  "forward",
		Long: `Forward a deployment to local`,
		Run: func(cmd *cobra.Command, args []string) {
			err := core.Forward(request)
			if err != nil {
				cmd.PrintErr(err)
			}
		},
	}
	cmd.Flags().StringVarP(&request.Namespace, "namespace", "n", client.GetDefaultNamespace(), "Namespace for forwarding")
	cmd.Flags().StringVar(&request.DeploymentName, "deployment", "", "Deployment Name")
	cmd.Flags().IntVar(&request.LocalPort, "localPort", 0, "Local Port")
	cmd.Flags().IntVar(&request.ContainerPort, "containerPort", 0, "Container Port")
	cmd.MarkFlagRequired("deployment")
	cmd.MarkFlagRequired("localPort")
	cmd.MarkFlagRequired("containerPort")
	return cmd
}
