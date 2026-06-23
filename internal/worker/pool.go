package worker

import (
	"context"
	"hash/fnv"

	"github.com/AntonioForYou/wallet-api/internal/domain"
	"github.com/google/uuid"
)

type Pool struct {
	size    int
	workers []*Worker
}

func NewPool(size int, bufferSize int, repo domain.WalletRepository) *Pool {
	workers := make([]*Worker, size)
	for i := 0; i < size; i++ {
		workers[i] = NewWorker(i, bufferSize, repo)
	}

	return &Pool{
		size:    size,
		workers: workers,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for _, w := range p.workers {
		w.Start(ctx)
	}
}

func (p *Pool) Dispatch(job domain.Job) {
	workerIndex := p.hashUUID(job.WalletID) % uint32(p.size)

	p.workers[workerIndex].jobChan <- job
}

func (p *Pool) hashUUID(id uuid.UUID) uint32 {
	h := fnv.New32a()
	h.Write(id[:])
	return h.Sum32()
}
