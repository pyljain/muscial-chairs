package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mc/internal/store"
	"net/http"
	"os"
	"path"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context, cmd *cli.Command) error {
	log.Printf("Starting server")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	kubeConfigPath := path.Join(homeDir, ".kube", "config")
	stopChan := make(chan struct{})
	log.Printf("Here 1")
	k8sStore, err := store.NewK8s(kubeConfigPath, stopChan)
	if err != nil {
		return err
	}
	log.Printf("Here 2")

	k8sStore.SetCallback(func() {
		err := reconcileWorkers(ctx, k8sStore, int(cmd.Int("worker-count")), int(cmd.Int("threshold")))
		if err != nil {
			log.Printf("Error reconciling workers: %v", err)
		}
	})

	err = reconcileWorkers(ctx, k8sStore, int(cmd.Int("worker-count")), int(cmd.Int("threshold")))
	if err != nil {
		return err
	}

	// Spin up an HTTP server and expose a route to the workers
	mux := http.NewServeMux()
	mux.HandleFunc("/busy", func(w http.ResponseWriter, r *http.Request) {
		wr := busyRequest{}
		err := json.NewDecoder(r.Body).Decode(&wr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = k8sStore.UpdateWorkerStatus(r.Context(), wr.WorkerID, "Running")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	})
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		return err
	}

	// Assign tasks to workers, update MongoDB with worker status "Running"

	// Wait for the worker to complete and update the status to "Complete"

	return nil
}

func reconcileWorkers(ctx context.Context, k8sStore store.DB, expectedPoolSize, poolThreshold int) error {

	log.Printf("Reconciling workers with expected pool size %d", expectedPoolSize)
	// Query to find number of workers that have status = 'Waiting'
	workers, err := k8sStore.GetWorkersByStatus(ctx, "Waiting")
	if err != nil {
		return err
	}

	log.Printf("Found %d waiting workers", len(workers))

	if len(workers) < expectedPoolSize {
		for i := 1; i <= expectedPoolSize-len(workers); i++ {

			// Generate uuid
			// id := uuid.New().String()

			id := petname.Generate(2, "-")
			workerName := fmt.Sprintf("mc-worker-%s", id)

			// Create Kubernetes Pod
			err := k8sStore.CreateWorker(ctx, workerName)
			if err != nil {
				return err
			}
		}
	} else if len(workers) > expectedPoolSize+poolThreshold {
		for _, w := range workers[expectedPoolSize+poolThreshold:] {
			err := k8sStore.DeleteWorker(ctx, w.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type busyRequest struct {
	WorkerID string `json:"worker_id"`
}
