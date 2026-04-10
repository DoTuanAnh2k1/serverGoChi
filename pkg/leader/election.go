// Package leader provides leader election using Kubernetes Lease API.
package leader

import (
	"context"
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
)

// Start runs the leader election loop until ctx is cancelled.
// onLeader is called when this pod becomes leader.
func Start(ctx context.Context, cfg config_models.LeaderConfig, onLeader func(ctx context.Context)) {
	client, err := buildK8sClient()
	if err != nil {
		logger.Logger.Errorf("leader: cannot build k8s client: %v", err)
		return
	}

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      cfg.LeaseName,
			Namespace: cfg.Namespace,
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: cfg.PodName,
		},
	}

	leaseDuration := duration(cfg.LeaseDurationSeconds, 15)
	renewDeadline := duration(cfg.RenewDeadlineSeconds, 10)
	retryPeriod := duration(cfg.RetryPeriodSeconds, 2)

	logger.Logger.Infof("leader: starting election — lease=%s/%s identity=%s duration=%v",
		cfg.Namespace, cfg.LeaseName, cfg.PodName, leaseDuration)

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   leaseDuration,
		RenewDeadline:   renewDeadline,
		RetryPeriod:     retryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(leaderCtx context.Context) {
				logger.Logger.Infof("leader: pod %q acquired lease — running leader tasks", cfg.PodName)
				onLeader(leaderCtx)
			},
			OnStoppedLeading: func() {
				logger.Logger.Errorf("leader: pod %q lost lease — restarting", cfg.PodName)
			},
			OnNewLeader: func(identity string) {
				if identity != cfg.PodName {
					logger.Logger.Infof("leader: current leader is %q (this pod is worker)", identity)
				}
			},
		},
	})
}

// buildK8sClient tries in-cluster config first, falls back to kubeconfig for local dev.
func buildK8sClient() (*kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			home, _ := os.UserHomeDir()
			kubeconfig = fmt.Sprintf("%s/.kube/config", home)
		}
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("build k8s config: %w", err)
		}
	}
	return kubernetes.NewForConfig(cfg)
}

func duration(seconds, defaultSeconds int) time.Duration {
	if seconds <= 0 {
		return time.Duration(defaultSeconds) * time.Second
	}
	return time.Duration(seconds) * time.Second
}
