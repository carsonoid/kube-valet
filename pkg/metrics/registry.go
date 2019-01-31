package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Registry is a struct to hold prometheus metrics
type Registry struct {
	packLeftPercentFullByNag map[string]*prometheus.GaugeVec
}

// NewRegistry returns a new Registry
func NewRegistry() *Registry {
	return &Registry{
		packLeftPercentFullByNag: make(map[string]*prometheus.GaugeVec),
	}
}

// GetPackLeftPercentFull initializes the metric map if needed and returns the result
func (r *Registry) GetPackLeftPercentFull(name string) *prometheus.GaugeVec {
	if _, ok := r.packLeftPercentFullByNag[name]; !ok {
		r.packLeftPercentFullByNag[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:        "kubevalet_packleft_full_percent",
			ConstLabels: prometheus.Labels{"node_assignment_group": name},
		}, []string{
			"node_assignment",
			"node_name",
			"pack_left_state",
		})
		prometheus.MustRegister(r.packLeftPercentFullByNag[name])
	}
	return r.packLeftPercentFullByNag[name]
}
