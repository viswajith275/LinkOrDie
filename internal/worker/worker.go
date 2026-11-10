package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"link_or_die/internal/apperror"
	"link_or_die/internal/checker"
	"link_or_die/internal/config"
	"link_or_die/internal/domains/links"
	"link_or_die/internal/lifecycle"
	"link_or_die/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Worker struct {
	pool     *pgxpool.Pool
	repo     *store.LinkRepository
	checker  *checker.Checker
	hosts    *hostLimiter
	log      *slog.Logger
	cfg      *config.Config
	lastTick atomic.Int64
}

func NewWorker(
	pool *pgxpool.Pool,
	repo *store.LinkRepository,
	urlChecker *checker.Checker,
	cfg *config.Config,
	logger *slog.Logger,
) *Worker {
	if logger == nil {
		logger = slog.Default()
	}

	return &Worker{
		pool:    pool,
		repo:    repo,
		checker: urlChecker,
		cfg:     cfg,
		hosts:   newHostLimiter(cfg.HostInterval),
		log:     logger,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if err := w.tick(ctx); err != nil {
				if !errors.Is(err, context.Canceled) {
					w.log.Error("worker tick failed", "err", err)
				}
			} else {
				w.lastTick.Store(time.Now().UnixNano())
			}
			timer.Reset(w.cfg.Interval)
		}
	}
}

func (w *Worker) Healthy(maxAge time.Duration) bool {
	if maxAge <= 0 {
		maxAge = 2 * time.Minute
	}

	last := w.lastTick.Load()
	if last == 0 {
		return false
	}

	return time.Since(time.Unix(0, last)) <= maxAge
}

func (w *Worker) tick(ctx context.Context) error {
	due, err := w.repo.ClaimDue(ctx, w.cfg.BatchSize, w.cfg.Lease)
	if err != nil {
		return fmt.Errorf("claim due links: %w", err)
	}
	if len(due) == 0 {
		return nil
	}

	sem := make(chan struct{}, w.cfg.Concurrency)
	var wg sync.WaitGroup

	for _, link := range due {
		link := link
		wg.Add(1)

		go func() {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				w.requeue(link, time.Now())
				return
			}

			w.process(ctx, link)
		}()
	}

	wg.Wait()
	return nil
}

func (w *Worker) process(ctx context.Context, link links.DueLink) {
	defer func() {
		if r := recover(); r != nil {
			w.log.Error("check panicked", "link_id", link.ID, "panic", r)
			w.requeue(link, time.Now().Add(time.Minute))
		}
	}()

	if delay := w.hosts.Reserve(link.Host); delay > 0 {
		w.requeue(link, time.Now().Add(delay))
		return
	}

	result := w.checker.Check(ctx, link.URL)

	if result.Canceled {
		w.requeue(link, time.Now())
		return
	}

	tx, err := w.pool.Begin(ctx)
	if err != nil {
		w.log.Error("begin check transaction", "link_id", link.ID, "err", err)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	checkedAt := time.Now()
	decision, err := lifecycle.Apply(ctx, tx, link.ID, result, checkedAt, link.LeaseUntil, w.cfg)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			w.requeue(link, time.Now())
		case errors.Is(err, lifecycle.ErrStaleClaim):
			// Another worker owns a newer claim; leave it alone.
		case errors.Is(err, apperror.ErrNotFound):
			// The link was deleted while it was being checked.
		default:
			w.log.Error("apply check result", "link_id", link.ID, "err", err)
		}
		return
	}

	if err := tx.Commit(ctx); err != nil {
		w.log.Error("commit check result", "link_id", link.ID, "err", err)
		return
	}

	w.log.Info("check applied",
		"link_id", link.ID,
		"host", link.Host,
		"outcome", result.Outcome,
		"state", decision.NextState,
		"changed", decision.Changed,
		"elapsed_ms", result.Elapsed.Milliseconds(),
	)
}

func (w *Worker) requeue(link links.DueLink, at time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := w.repo.Requeue(ctx, link.ID, link.LeaseUntil, at); err != nil {
		w.log.Error("requeue link", "link_id", link.ID, "err", err)
	}
}
