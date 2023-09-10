/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kbudde/k8n/internal/ytt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// localCmd represents the local command
//
//nolint:exhaustruct,gochecknoglobals,
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Render the ytt templates locally",
	Long:  `Takes the input file and the ytt files and renders the templates locally.`,
	Run: func(cmd *cobra.Command, args []string) {
		var outW io.Writer
		var file *os.File

		inputFile := viper.GetString("input")
		folder := viper.GetString("ytt")

		if folder == "" {
			folder = filepath.Dir(inputFile)
		}
		_, err := os.ReadFile(inputFile)
		cobra.CheckErr(err)

		outputFile := viper.GetString("output")

		if outputFile == "-" {
			outW = os.Stdout
		} else {
			file, err = os.Create(outputFile)
			cobra.CheckErr(err)
			outW = file
		}
		defer file.Close()

		out, err := ytt.Render(inputFile, folder)
		if err != nil {
			fmt.Fprintln(os.Stderr, string(out))
			cobra.CheckErr(err)
		}

		fmt.Fprintln(outW, string(out))
	},
}

func init() {
	renderCmd.AddCommand(localCmd)
	localCmd.Flags().String("input", "input.yaml", "path to data file. See `k8n read`")
	localCmd.Flags().String("ytt", "", "path to ytt files. Defaults to the directory of the input file.")

	err := viper.BindPFlags(localCmd.Flags())
	cobra.CheckErr(err)
}
