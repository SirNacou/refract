package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/SirNacou/refract/api/internal/config"
	ingestclicks "github.com/SirNacou/refract/api/internal/features/clicks/ingest_clicks"
	"github.com/SirNacou/refract/api/internal/infrastructure/publisher"
	"github.com/valkey-io/valkey-go"
)

type ClicksStreamWorker struct {
	valkey   valkey.Client
	handler  *ingestclicks.CommandHandler
	cfg      *config.ValkeyConfig
	stopChan chan error
}

type Batch struct {
	Clicks []ingestclicks.Click
	IDs    []string
}

func NewClicksStreamWorker(ctx context.Context, valkey valkey.Client, handler *ingestclicks.CommandHandler, cfg *config.ValkeyConfig) (*ClicksStreamWorker, error) {
	cmd := valkey.B().XgroupCreate().Key(cfg.ClicksStreamKey).Group(cfg.ReadGroup).Id("0").Mkstream().Build()
	err := valkey.Do(ctx, cmd).Error()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return nil, err
	}

	return &ClicksStreamWorker{
		valkey:   valkey,
		handler:  handler,
		cfg:      cfg,
		stopChan: make(chan error),
	}, nil
}

func (w *ClicksStreamWorker) Start(ctx context.Context) error {

	go func() {
		w.stopChan <- w.loop(ctx)
	}()

	return nil
}

func (w *ClicksStreamWorker) Stop(ctx context.Context) error {
	select {
	case err := <-w.stopChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *ClicksStreamWorker) loop(ctx context.Context) error {
	batch := &Batch{}
	ticker := time.NewTicker(time.Second * 5)
	pendingTicker := time.NewTicker(time.Minute * 1)
	defer ticker.Stop()
	for {
		wCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		// Read NEW messages (block here)
		err := w.readNewClicks(ctx, batch)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to read new", "error", err)
			time.Sleep(time.Second)
		}

		select {
		case <-wCtx.Done():
			if len(batch.Clicks) > 0 {
				w.flush(wCtx, batch)
			}
			return wCtx.Err()
		case <-ticker.C:
			if len(batch.Clicks) > 0 {
				w.flush(wCtx, batch)
			}
		default:
			if len(batch.Clicks) > w.cfg.BatchSize {
				w.flush(wCtx, batch)
			}
		}

		select {
		case <-pendingTicker.C:
			err = w.readPendingClicks(ctx, batch)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to read pending", "error", err)
			}
		default:
		}
	}
}

func (w *ClicksStreamWorker) readNewClicks(ctx context.Context, batch *Batch) error {
	newCmd := w.valkey.B().Xreadgroup().
		Group(w.cfg.ReadGroup, w.cfg.Consumer).
		Count(100).
		Block((time.Second * 5).Milliseconds()).
		Streams().
		Key(w.cfg.ClicksStreamKey).
		Id(">"). // New messages
		Build()

	newRes, err := w.valkey.Do(ctx, newCmd).AsXRead()
	if err != nil && !valkey.IsValkeyNil(err) {
		return err
	} else if newStream, ok := newRes[w.cfg.ClicksStreamKey]; ok && len(newStream) > 0 {
		slog.Info("Found new messages", "count", len(newStream))
		items, ids, err := w.processBatch(ctx, newStream)
		if err == nil {
			batch.Clicks = append(batch.Clicks, items...)
			batch.IDs = append(batch.IDs, ids...)
		}
	}

	return nil
}

func (w *ClicksStreamWorker) readPendingClicks(ctx context.Context, batch *Batch) error {
	pendingCmd := w.valkey.B().Xreadgroup().
		Group(w.cfg.ReadGroup, w.cfg.Consumer).
		Count(100).
		Streams().
		Key(w.cfg.ClicksStreamKey).
		Id("0"). // Pending
		Build()  // No Block() here!

	pendingRes, err := w.valkey.Do(ctx, pendingCmd).AsXRead()
	if err != nil && !valkey.IsValkeyNil(err) {
		return err
	} else if pendingStream, ok := pendingRes[w.cfg.ClicksStreamKey]; ok && len(pendingStream) > 0 {
		slog.Info("Found pending messages", "count", len(pendingStream))
		items, ids, err := w.processBatch(ctx, pendingStream)
		if err == nil {
			batch.Clicks = append(batch.Clicks, items...)
			batch.IDs = append(batch.IDs, ids...)
		}
	}

	return nil
}

func (w *ClicksStreamWorker) processBatch(ctx context.Context, batch []valkey.XRangeEntry) (clicks []ingestclicks.Click, ids []string, err error) {
	for _, v := range batch {
		data, ok := v.FieldValues["data"]
		if !ok {
			continue
		}

		slog.Info("Message", "id", v.ID, "value", v.FieldValues)

		req := &publisher.ClicksPublisherRequest{}
		err := json.Unmarshal([]byte(data), &req)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to unmarshal click", "error", err)
			continue
		}

		clicks = append(clicks, ingestclicks.Click{
			ShortCode: req.ShortCode,
			ClickedAt: req.ClickedAt,
			IPAddress: req.IPAddress,
			UserAgent: req.UserAgent,
			Referer:   req.Referer,
		})
		ids = append(ids, v.ID)
	}

	return clicks, ids, nil
}

func (w *ClicksStreamWorker) flush(ctx context.Context, batch *Batch) error {
	err := w.handler.Handle(ctx, &ingestclicks.Command{
		Clicks: batch.Clicks,
	})

	if err != nil {
		slog.ErrorContext(ctx, "Failed to handle clicks batch", "error", err)
		return err
	}

	ackCmd := w.valkey.B().Xack().
		Key(w.cfg.ClicksStreamKey).
		Group(w.cfg.ReadGroup).
		Id(batch.IDs...).
		Build()

	err = w.valkey.Do(ctx, ackCmd).Error()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to acknowledge clicks", "error", err)
		return err
	}

	batch.Clicks = []ingestclicks.Click{}
	batch.IDs = []string{}

	return nil
}
