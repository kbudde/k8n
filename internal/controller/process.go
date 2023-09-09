package controller

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

//nolint:gochecknoglobals
var (
	watchedResources = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "k8n_watched_resources",
		Help: "Number of watched resources",
	}, []string{"name"})
	syncOperations_total = promauto.NewCounter(prometheus.CounterOpts{
		Name: "k8n_sync_operations_total",
		Help: "Number of sync operations",
	})
)

func (c *Controller) doTheStuff(key string) error {
	fmt.Printf("Change detected '%s'\n", key)
	syncOperations_total.Inc()

	data := map[string]interface{}{}

	watchedResources.Reset()

	for name, index := range c.indexer {
		items := index.List()

		// Create a slice to store the modified items
		modifiedItems := []interface{}{}
		// Remove the "object" key from each item
		for _, item := range items {
			unstructuredItem, ok := item.(*unstructured.Unstructured)
			if !ok {
				return fmt.Errorf("error: item is not of type unstructured.Unstructured")
			}

			modifiedItems = append(modifiedItems, unstructuredItem.Object)
			watchedResources.WithLabelValues(name).Set(float64(len(modifiedItems)))
		}

		data[name] = modifiedItems
	}

	// Convert the data to YAML
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling data to YAML: %w", err)
	}

	c.process <- yamlData

	return nil
}
