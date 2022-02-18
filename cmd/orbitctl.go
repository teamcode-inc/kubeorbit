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
package main

import (
	"github.com/spf13/cobra"
	"kubeorbit.io/pkg/cli/command"
)

var rootCmd = &cobra.Command{
	Use:  "orbitctl",
	Long: `Orbitctl can forward traffic intended for a service in-clsuter to your local workload`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Example: "orbitctl forward --deployment depolyment-a --namespace ns-a --containerPort 8080 --localPort 8080",
	Version: "0.2.0",
}

func init() {
	rootCmd.AddCommand(command.ForwardCommand())
	rootCmd.AddCommand(command.UninstallCommand())
}

func main() {
	cobra.CheckErr(rootCmd.Execute())
}
