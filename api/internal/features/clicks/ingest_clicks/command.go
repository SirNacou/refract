package ingestclicks

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Command struct {
	Clicks []Click
}

type Click struct {
	ShortCode string
	ClickedAt time.Time // ← Parse from message
	IPAddress string
	UserAgent string
	Referer   string
}

type CommandHandler struct {
	chClient driver.Conn
}

func NewCommandHandler(conn driver.Conn) *CommandHandler {
	return &CommandHandler{
		chClient: conn,
	}
}

func (h *CommandHandler) Handle(ctx context.Context, cmd *Command) error {
	batch, err := h.chClient.PrepareBatch(ctx, "INSERT INTO clicks (short_code, clicked_at, ip_address, user_agent, referer)")
	if err != nil {
		return err
	}
	defer batch.Close()

	for _, click := range cmd.Clicks {
		batch.Append(
			click.ShortCode,
			click.ClickedAt,
			click.IPAddress,
			click.UserAgent,
			click.Referer,
		)
	}

	return batch.Send()
}
