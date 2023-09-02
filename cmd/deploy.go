/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/kbudde/k8n/internal/kapp"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command.
//
//nolint:exhaustruct,gochecknoglobals
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy us kapp",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		folder, err := cmd.Flags().GetString("folder")
		cobra.CheckErr(err)
		k8s, err := kubeConfigFromFlags()
		cobra.CheckErr(err)

		kapp := kapp.New(k8s)
		out, err := kapp.Deploy("TODO", folder)
		fmt.Println(out)
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	// folder flag
	deployCmd.Flags().StringP("folder", "f", "", "folder where the kubernetes manifests are located")
}
