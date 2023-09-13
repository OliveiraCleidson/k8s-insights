/*
Copyright © 2023 Cleidson Oliveira <contato@cleidsonoliveira.dev>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kin",
	Short: "A simple CLI that extract data from Kubernetes cluster and provide insights about resource utilization. ",
	Long:  `This CLI extract data from Kubernetes cluster and provide information like startup time of deployments, resource utilization of pods, etc.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
