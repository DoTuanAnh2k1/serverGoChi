// Package leader cung cấp cơ chế leader election sử dụng Kubernetes Lease API.
// Mỗi pod chạy election loop; pod nào giữ được Lease thì trở thành leader
// và được phép chạy các tác vụ đặc quyền (export CSV, v.v.).
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

	"github.com/DoTuanAnh2k1/serverGoChi/internal/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
)

// Start chạy leader election loop cho đến khi ctx bị cancel.
// onLeader được gọi (trong goroutine riêng) khi pod này trở thành leader;
// context truyền vào onLeader sẽ bị cancel khi pod mất lease.
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

// buildK8sClient thử in-cluster config trước, fallback sang kubeconfig cho local dev.
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
