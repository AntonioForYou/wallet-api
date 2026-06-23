package worker

import (
	"context"
	"log"

	"github.com/AntonioForYou/wallet-api/internal/domain"
)

type Worker struct {
	id      int
	jobChan chan domain.Job
	repo    domain.WalletRepository
}

func NewWorker(id int, bufferSize int, repo domain.WalletRepository) *Worker {
	return &Worker{
		id:      id,
		jobChan: make(chan domain.Job, bufferSize),
		repo:    repo,
	}
}

func (w *Worker) Start(ctx context.Context) {
	go func() {
		log.Printf("Worker %d started\n", w.id)
		for {
			select {
			case <-ctx.Done():
				log.Printf("Worker %d stopping\n", w.id)
				return
			case job := <-w.jobChan:
				w.processJob(job)
			}
		}
	}()
}

func (w *Worker) processJob(job domain.Job) {
	amount := job.Amount
	if job.OperationType == domain.Withdraw {
		amount = -amount
	}

	newBalance, err := w.repo.UpdateBalance(job.Ctx, job.WalletID, amount)

	select {
	case job.ResultChan <- domain.Result{NewBalance: newBalance, Err: err}:
	case <-job.Ctx.Done():
		log.Printf("Worker %d: client disconnected or timeout for wallet %s\n", w.id, job.WalletID)
	}
}
