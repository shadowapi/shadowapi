package handler

import (
	"context"
	"net/http"

	"github.com/shadowapi/shadowapi/backend/internal/worker"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

func (h *Handler) WorkerJobsCancel(ctx context.Context, params api.WorkerJobsCancelParams) error {
	log := h.log.With("handler", "WorkerJobsCancel", "uuid", params.UUID)

	if !worker.CancelJob(params.UUID) {
		log.Error("job not found or already finished")
		return ErrWithCode(http.StatusNotFound, E("job not found or already running"))
	}

	log.Info("cancellation signaled")
	// returning nil ⇒ 204 No Content
	return nil
}
