/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kbudde/k8n/internal/config"
	"github.com/kbudde/k8n/internal/controller"
	"github.com/kbudde/k8n/internal/kapp"
	"github.com/kbudde/k8n/internal/processor"
	"github.com/kbudde/k8n/internal/ytt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command.
//
//nolint:exhaustruct,gochecknoglobals,lll
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run",
	Long:  `Run continuously, reading the data from the cluster and rendering the templates and applying them.`,
	Run: func(cmd *cobra.Command, args []string) {
		osSignal := make(chan os.Signal, 1)
		signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)
		server := &http.Server{
			Addr: ":59712",
		}
		server.Handler = promhttp.Handler()
		go func() {
			log.Println("Starting server...")
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				cobra.CheckErr(err)
			}
		}()
		configFile := viper.GetString("config")

		fmt.Printf("configFile: %v\n", configFile)
		cfg, err := config.FromYamlFile(configFile)
		cobra.CheckErr(err)

		folder := viper.GetString("ytt")
		if folder == "" {
			folder = filepath.Dir(configFile)
		}
		restConfig, err := kubeConfigFromFlags()
		cobra.CheckErr(err)

		ctrl, err := controller.NewForConfig(*cfg, restConfig)
		cobra.CheckErr(err)
		if ctrl == nil {
			panic("controller is nil")
		}
		stop := make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-osSignal
			timeout, cancelT := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelT()
			server.Shutdown(timeout)
			close(stop)
			cancel()
		}()
		go func() {
			ctrl.Run(stop)
		}()

		kapp := kapp.New()

		reconcile := viper.GetDuration("reconcile")

		p := processor.Processor{
			Controller:    *ctrl,
			Deployer:      kapp,
			RenderFunc:    ytt.Render,
			Name:          cfg.Metadata.Name,
			Folder:        folder,
			ReconcileTime: reconcile,
		}
		err = p.Process(ctx)
		cobra.CheckErr(err)

		fmt.Println("Done")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().String("config", "config.yaml", "path to config file.")
	runCmd.Flags().String("ytt", "", "path to ytt files. Defaults to the directory of the input file.")
	runCmd.Flags().Duration("reconcile", 30*time.Second, "reconcile time between consecutive syncs.")

	err := viper.BindPFlags(runCmd.Flags())
	cobra.CheckErr(err)
}
