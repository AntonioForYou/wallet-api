package worker

import (
	"context"
	"errors"
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
	var amount int64

	switch job.OperationType {
	case domain.Deposit:
		amount = job.Amount
	case domain.Withdraw:
		amount = -job.Amount
	default:
		w.sendResult(job, domain.Result{Err: errors.New("unknown operation type")})
		return
	}

	newBalance, err := w.repo.UpdateBalance(job.Ctx, job.WalletID, amount)

	w.sendResult(job, domain.Result{NewBalance: newBalance, Err: err})
}

func (w *Worker) sendResult(job domain.Job, result domain.Result) {
	select {
	case job.ResultChan <- result:
	case <-job.Ctx.Done():
		log.Printf("Worker %d: client disconnected or timeout for wallet %s\n", w.id, job.WalletID)
	}
}
