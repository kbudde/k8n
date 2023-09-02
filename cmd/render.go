/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// renderCmd represents the render command.
//
//nolint:exhaustruct,gochecknoglobals,lll
var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render templates without applying",
	Long:  `Retrieve data from cluster and render locally the ytt templates. The data from the cluster is provided as input.`,
}

func init() {
	rootCmd.AddCommand(renderCmd)
	renderCmd.PersistentFlags().String("output", "-", "output the file for data of the cluster, default is stdout")
}
