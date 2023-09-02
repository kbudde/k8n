package kapp

import (
	"os"
	"os/exec"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Deployer interface {
	Deploy(name string, folder string) (string, error)
}

type Kapp struct {
	config *rest.Config
}

// New returns a new kapp deployer.
func New(client *rest.Config) Deployer {
	return &Kapp{
		config: client,
	}
}

// Deploys a kubernetes application using kapp.
// name is the name of the application.
// manifest is the kubernetes manifests.
func (k *Kapp) Deploy(name string, manifest string) (string, error) {
	kube, cleanup, err := writeKubeconfig(*k.config)
	if err != nil {
		return "", err
	}

	defer cleanup()

	args := []string{
		"deploy",
		"-a", name,
		"-f", manifest,
		"--kubeconfig", kube,
		"--dangerous-allow-empty-list-of-resources",
		"-y",
	}

	return runKapp(args)
}

// Run kapp with the given args.
func runKapp(args []string) (string, error) {
	cmd := exec.Command("kapp", args...)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

func writeKubeconfig(config rest.Config) (string, func(), error) {
	kubeconfig := createKubeconfigFileForRestConfig(config)
	cleanup := func() {
		os.Remove(kubeconfig)
	}

	return kubeconfig, cleanup, nil
}

func createKubeconfigFileForRestConfig(restConfig rest.Config) string {
	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters["default-cluster"] = &clientcmdapi.Cluster{
		Server:                   restConfig.Host,
		CertificateAuthorityData: restConfig.CAData,
	}
	contexts := make(map[string]*clientcmdapi.Context)
	contexts["default-context"] = &clientcmdapi.Context{
		Cluster:  "default-cluster",
		AuthInfo: "default-user",
	}
	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos["default-user"] = &clientcmdapi.AuthInfo{
		ClientCertificateData: restConfig.CertData,
		ClientKeyData:         restConfig.KeyData,
	}
	clientConfig := clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: "default-context",
		AuthInfos:      authinfos,
	}
	kubeConfigFile, _ := os.CreateTemp("", "kubeconfig")
	_ = clientcmd.WriteToFile(clientConfig, kubeConfigFile.Name())

	return kubeConfigFile.Name()
}
