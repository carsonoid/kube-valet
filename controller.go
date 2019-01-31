package main

import (
	"github.com/op/go-logging"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"

	valetclient "github.com/domoinc/kube-valet/pkg/client/clientset/versioned"

	"github.com/domoinc/kube-valet/pkg/config"
	"github.com/domoinc/kube-valet/pkg/controller"
)

// KubeValet holds the basic config and clients
type KubeValet struct {
	kubeClientset  *kubernetes.Clientset
	valetClientset *valetclient.Clientset
	stopChan       chan struct{}
	config         *config.ValetConfig
}

// NewKubeValet takes clients and config and returns a KubeValet controller
func NewKubeValet(kc *kubernetes.Clientset, dc *valetclient.Clientset, config *config.ValetConfig) *KubeValet {
	logging.SetBackend(config.LoggingBackend)
	return &KubeValet{
		kubeClientset:  kc,
		valetClientset: dc,
		config:         config,
	}
}

// StartControllers starts the Kubernetes resource processing controllers
func (kd *KubeValet) StartControllers(stop <-chan struct{}) {
	resourceWatcher := controller.NewResourceWatcher(kd.kubeClientset, kd.valetClientset, kd.config)
	resourceWatcher.Run(kd.stopChan)
}

// StopControllers starts the Kubernetes resource processing controllers
func (kd *KubeValet) StopControllers() {
	close(kd.stopChan)
}

// Run starts the KubeValet instance and blocks until stopped
func (kd *KubeValet) Run() {
	// Create a channel for leader elect events
	// and exit signaling
	kd.stopChan = make(chan struct{})
	defer close(kd.stopChan)

	log.Notice("Running controllers")
	// Do Election
	if *leaderElection {
		log.Debug("Leader election enabled")

		log.Debug("Building ResourceLock")
		rl, err := resourcelock.New(
			*electResource,
			*electNamespace,
			*electName,
			kd.kubeClientset.CoreV1(),
			resourcelock.ResourceLockConfig{
				Identity: *electID,
				EventRecorder: record.NewBroadcaster().NewRecorder(
					scheme.Scheme,
					corev1.EventSource{Component: KubernetesComponent},
				),
			},
		)
		if err != nil {
			log.Fatalf("Error building ResourceLock: %s", err)
		}

		log.Debug("Building LeaderElector")

		leaderelection.RunOrDie(leaderelection.LeaderElectionConfig{
			Lock:          rl,
			LeaseDuration: *electDuration,
			RenewDeadline: *electDeadline,
			RetryPeriod:   *electRetry,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: kd.StartControllers,
				OnStoppedLeading: kd.StopControllers,
				OnNewLeader: func(identity string) {
					log.Debugf("Observed %s as the leader", identity)
				},
			},
		})
	} else {
		log.Debug("Leader election disabled. Running controllers.")
		kd.StartControllers(kd.stopChan)
		<-kd.stopChan
	}
}
