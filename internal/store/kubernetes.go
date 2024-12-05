package store

import (
	"context"
	"encoding/json"
	"fmt"
	"mc/internal/models"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	corev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type k8s struct {
	clientset   *kubernetes.Clientset
	podInformer corev1.PodInformer
}

const (
	mcNamespace = "musicalchairs"
)

func NewK8s(kubeconfig string, stopCh chan struct{}) (DB, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	selector := labels.SelectorFromSet(map[string]string{"app": "mc-worker"})
	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		time.Hour*24,
		informers.WithNamespace("musicalchairs"),
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = selector.String()
		}),
	)

	podInformer := informerFactory.Core().V1().Pods()

	go func() {
		informerFactory.Start(stopCh)
	}()

	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, podInformer.Informer().HasSynced) {
		return nil, err
	}

	return &k8s{clientset: clientset, podInformer: podInformer}, nil
}

func (db *k8s) SetCallback(callback func()) {
	db.podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			workerPod := obj.(*v1.Pod)
			if workerPod.Status.Phase != v1.PodRunning {
				callback()
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			callback()
		},
		DeleteFunc: func(obj interface{}) {
			callback()
		},
	})
}

func (db *k8s) GetWorkersByStatus(ctx context.Context, status string) ([]models.Worker, error) {

	selector := labels.SelectorFromSet(map[string]string{"status": status, "app": "mc-worker"})
	pods, err := db.podInformer.Lister().List(selector)
	if err != nil {
		return nil, err
	}

	eligibleWorkers := []models.Worker{}

	for _, pod := range pods {
		eligibleWorkers = append(eligibleWorkers, models.Worker{
			ID:     pod.Name,
			Status: pod.Labels["status"],
		})
	}

	return eligibleWorkers, nil
}

func (db *k8s) CreateWorker(ctx context.Context, workerName string) error {
	podsClient := db.clientset.CoreV1().Pods(mcNamespace)

	workerPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: workerName,
			Labels: map[string]string{
				"app":    "mc-worker",
				"status": "Waiting",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "mc-worker",
					Image: "alpine",
					Command: []string{
						"sleep",
						"3000",
					},
				},
			},
		},
	}

	fmt.Printf("Creating pod %s\n", workerName)
	_, err := podsClient.Create(ctx, workerPod, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (db *k8s) UpdateWorkerStatus(ctx context.Context, podname, status string) error {
	podsClient := db.clientset.CoreV1().Pods(mcNamespace)

	patchData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				"status": status,
			},
		},
	}

	// Convert the patch data to JSON
	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		return err
	}

	_, err = podsClient.Patch(ctx, podname, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (db *k8s) DeleteWorker(ctx context.Context, podName string) error {
	err := db.clientset.CoreV1().Pods(mcNamespace).Delete(ctx, podName, *&metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
