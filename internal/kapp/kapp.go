package kapp

import (
	"os/exec"
)

type Deployer interface {
	Deploy(name string, folder string) (string, error)
}

type Kapp struct {
}

// New returns a new kapp deployer.
func New() Deployer {
	return &Kapp{}
}

// Deploys a kubernetes application using kapp.
// name is the name of the application.
// manifest is the kubernetes manifests.
func (k *Kapp) Deploy(name string, manifest string) (string, error) {
	args := []string{
		"deploy",
		"-a", name,
		"-f", manifest,
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
