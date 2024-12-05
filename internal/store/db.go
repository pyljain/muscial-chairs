package store

import (
	"context"
	"mc/internal/models"
)

type DB interface {
	GetWorkersByStatus(ctx context.Context, status string) ([]models.Worker, error)
	CreateWorker(ctx context.Context, id string) error
	UpdateWorkerStatus(ctx context.Context, id, status string) error
	DeleteWorker(ctx context.Context, podName string) error
	SetCallback(callback func())
}
