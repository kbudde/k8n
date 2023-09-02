/*
Copyright © 2023 Kris Budde

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	goflags "flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

//nolint:gochecknoglobals
var (
	cfgFile        string
	kubeconfigPath string
)

// rootCmd represents the base command when called without any subcommands.
//
//nolint:exhaustruct,gochecknoglobals
var rootCmd = &cobra.Command{
	Use:   "k8n",
	Short: "Manage resources in Kubernetes with the power of k8n (Kuberian)",
	//nolint:lll
	Long: `k8n (pronounced "ken") is a command-line tool that simplifies the management of resources in Kubernetes clusters by leveraging the power of a universal operator. 
	       This operator generates resources based on existing resources in the cluster, and can be configured to watch resources based on API version, kind, and labels/annotations.
				 On each change, the resources are provided as input to ytt overlays, which transform the resources into Kubernetes manifests. The manifests are then applied to the cluster with kapp.

Examples:
- For each namespace with labels "team=something" and "feature=alerting", 
  k8n can generate Alertmanager configuration and apply it.
- For each deployed RabbitMQ pod, k8n will create a deployment for "kbudde/rabbitmq-exporter" to monitor the pod.

Use k8n to harness the magic of Kubernetes and effortlessly manage your containerized applications.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	fs := goflags.NewFlagSet("", goflags.PanicOnError)
	klog.InitFlags(fs)
	rootCmd.PersistentFlags().AddGoFlagSet(fs)

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .k8n.yaml)")
	rootCmd.PersistentFlags().StringVar(
		&kubeconfigPath,
		"kubeconfig",
		os.Getenv("KUBECONFIG"),
		"defaults to $KUBECONFIG if not in cluster")
}

func kubeConfigFromFlags() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err == nil {
		return cfg, nil
	}

	// If not running in-cluster, use the default kubeconfig file
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigType("yaml")
		viper.SetConfigName(".k8n")
	}

	viper.SetEnvPrefix("K8N")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func SetVersion(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf(`%s
Commit %s
BuildDate %s
Go %s`, version, commit, date, runtime.Version())
}
