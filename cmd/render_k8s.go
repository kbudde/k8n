/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kbudde/k8n/internal/config"
	"github.com/kbudde/k8n/internal/controller"
	"github.com/kbudde/k8n/internal/ytt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// k8sCmd represents the k8s command
//
//nolint:exhaustruct,gochecknoglobals
var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Render the ytt templates with data from k8s",
	Long:  `Takes the input file and the ytt files and renders the templates with data from k8s.`,
	Run: func(cmd *cobra.Command, args []string) {
		var outW io.Writer
		var file *os.File

		configFile, err := cmd.Flags().GetString("config")
		cobra.CheckErr(err)
		cfg, err := config.FromYamlFile(configFile)
		cobra.CheckErr(err)
		folder, err := cmd.Flags().GetString("ytt")
		cobra.CheckErr(err)
		if folder == "" {
			folder = filepath.Dir(configFile)
		}
		restConfig, err := kubeConfigFromFlags()
		cobra.CheckErr(err)
		data, err := controller.Read(restConfig, *cfg)
		cobra.CheckErr(err)
		yamlData, err := yaml.Marshal(data)
		cobra.CheckErr(err)

		// write yamlData to a temporary file
		tempFile, err := os.CreateTemp("", "input.yaml")
		cobra.CheckErr(err)
		defer os.Remove(tempFile.Name()) // clean up

		_, err = tempFile.Write(yamlData)
		cobra.CheckErr(err)
		defer tempFile.Close()

		out, err := ytt.Render(tempFile.Name(), folder)
		cobra.CheckErr(err)

		outputFile, err := cmd.Flags().GetString("output")
		cobra.CheckErr(err)

		if outputFile == "-" {
			outW = os.Stdout
		} else {
			file, err = os.Create(outputFile)
			cobra.CheckErr(err)
			outW = file
		}
		defer file.Close()

		fmt.Fprintln(outW, string(out))
	},
}

func init() {
	renderCmd.AddCommand(k8sCmd)
	k8sCmd.Flags().String("config", "config.yaml", "path to config file.")
	k8sCmd.Flags().String("ytt", "", "path to ytt files. Defaults to the directory of the config file.")
}
