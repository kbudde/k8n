package controller

import (
	"context"

	"github.com/kbudde/k8n/internal/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func Read(kube *rest.Config, cfg config.Config) (map[string]interface{}, error) {
	dynamicClient, err := dynamic.NewForConfig(kube)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}

	for _, watcher := range cfg.Watcher {
		gv, err := schema.ParseGroupVersion(watcher.APIVersion)
		if err != nil {
			return nil, err
		}

		gvr := gv.WithResource(watcher.Kind)
		//nolint:exhaustruct
		options := metav1.ListOptions{}
		options.LabelSelector = watcher.Selector
		options.APIVersion = watcher.APIVersion
		options.Kind = watcher.Kind

		items, err := dynamicClient.Resource(gvr).Namespace(watcher.Namespace).List(context.Background(), options)
		if err != nil {
			return nil, err
		}

		// Create a slice to store the modified items
		modifiedItems := []interface{}{}

		// Remove the "object" key from each item
		for _, item := range items.Items {
			modifiedItems = append(modifiedItems, item.Object)
		}

		data[watcher.Name] = modifiedItems
	}

	return data, nil
}
