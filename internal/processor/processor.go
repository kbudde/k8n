package processor

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kbudde/k8n/internal/controller"
	"github.com/kbudde/k8n/internal/kapp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/klog/v2"
)

// Prometheus metrics.
//
//nolint:gochecknoglobals
var (
	processingTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "k8n_processing_time_seconds",
		Help:    "Processing time in seconds",
		Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120, 300},
	},
		[]string{"step"})
	totalChanges = promauto.NewCounter(prometheus.CounterOpts{
		Name: "k8n_changes_total",
		Help: "Number of total changes detected and processed",
	})
	failedChanges = promauto.NewCounter(prometheus.CounterOpts{
		Name: "k8n_failed_changes_total",
		Help: "Number of failed changes detected and processed",
	})
	status = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "k8n_status",
		Help: "Status of processing steps",
	},
		[]string{"step"})
)

type Processor struct {
	Controller controller.Controller
	Deployer   kapp.Deployer
	RenderFunc func(input, folder string) ([]byte, error)
	Name       string
	Folder     string
}

func (p *Processor) Process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data := <-p.Controller.GetProcessChan():
			totalChanges.Inc()

			err := p.process(data)
			if err != nil {
				failedChanges.Inc()

				return err
			}
		}
	}
}

func (p *Processor) process(data []byte) error {
	now := time.Now()

	inputFile, err := p.TempFile(data)
	if err != nil {
		status.WithLabelValues("saveInput").Set(0)

		return err
	}

	defer p.Cleanup(inputFile)
	status.WithLabelValues("saveInput").Set(1)
	processingTime.WithLabelValues("saveInput").Observe(time.Since(now).Seconds())

	// Render the templates
	now = time.Now()

	rendered, err := p.RenderFunc(inputFile, p.Folder)
	if err != nil {
		status.WithLabelValues("render").Set(0)

		return fmt.Errorf("error rendering: %w\nout:%s\n", err, rendered)
	}

	status.WithLabelValues("render").Set(1)
	processingTime.WithLabelValues("render").Observe(time.Since(now).Seconds())

	// Create a temporary file for the rendered templates
	now = time.Now()

	manifestFile, err := p.TempFile(rendered)
	if err != nil {
		status.WithLabelValues("saveRendered").Set(0)

		return err
	}

	defer p.Cleanup(manifestFile)
	status.WithLabelValues("saveRendered").Set(1)
	processingTime.WithLabelValues("saveRendered").Observe(time.Since(now).Seconds())

	// Deploy the rendered templates
	now = time.Now()

	out, err := p.Deployer.Deploy(p.Name, manifestFile)
	if err != nil {
		klog.Errorf("error deploying: %v\nout: %s", err, out)
		status.WithLabelValues("deploy").Set(0)

		return nil
	}

	status.WithLabelValues("deploy").Set(1)
	processingTime.WithLabelValues("deploy").Observe(time.Since(now).Seconds())

	fmt.Printf("Deployed successfully.\n%s\n", out)

	return nil
}

func (p *Processor) TempFile(data []byte) (string, error) {
	tempFile, err := os.CreateTemp("", "manifest.yaml")
	if err != nil {
		return "", err
	}

	_, err = tempFile.Write(data)
	if err != nil {
		return "", err
	}

	err = tempFile.Close()
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

func (p *Processor) Cleanup(file string) error {
	return os.Remove(file)
}
