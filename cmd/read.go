/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/kbudde/k8n/internal/config"
	"github.com/kbudde/k8n/internal/controller"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// readCmd represents the read command.
//
//nolint:exhaustruct,gochecknoglobals
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read data from kubernetes cluster",
	Long: `Read the data from kubernetes cluster using the configuration.
	This is useful when you want to test the rendering locally.`,
	Run: func(cmd *cobra.Command, args []string) {
		var out io.Writer
		var file *os.File

		restConfig, err := kubeConfigFromFlags()
		cobra.CheckErr(err)
		cfgFile, err := cmd.Flags().GetString("config")
		cobra.CheckErr(err)
		output, err := cmd.Flags().GetString("output")
		cobra.CheckErr(err)
		cfg, err := config.FromYamlFile(cfgFile)
		cobra.CheckErr(err)
		data, err := controller.Read(restConfig, *cfg)
		cobra.CheckErr(err)
		// Convert the data to YAML
		yamlData, err := yaml.Marshal(data)
		cobra.CheckErr(err)

		if output == "-" {
			out = os.Stdout
		} else {
			file, err = os.Create(output)
			cobra.CheckErr(err)
			out = file
		}
		defer file.Close()

		fmt.Fprintln(out, string(yamlData))
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.Aliases = []string{"r"}
	readCmd.Flags().String("config", "config.yaml", "path to the configuration file")
	readCmd.Flags().String("output", "-", "output the file for data of the cluster, default is stdout")
}
