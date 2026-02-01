package publisher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/valkey-io/valkey-go"
)

type ClicksPublisher struct {
	valkey          valkey.Client
	clicksStreamKey string
}

func NewClicksPublisher(valkey valkey.Client, clicksStreamKey string) *ClicksPublisher {
	return &ClicksPublisher{
		valkey:          valkey,
		clicksStreamKey: clicksStreamKey,
	}
}

type ClicksPublisherRequest struct {
	ShortCode string    `json:"short_code"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Referer   string    `json:"referer,omitempty"`
	ClickedAt time.Time `json:"clicked_at"`
}

func (ct *ClicksPublisher) Publish(ctx context.Context, req *ClicksPublisherRequest) error {
	res, err := json.Marshal(req)
	if err != nil {
		return err
	}

	cmd := ct.valkey.B().Xadd().
		Key(ct.clicksStreamKey).
		Id("*").
		FieldValue().
		FieldValue("data", string(res)).
		Build()

	err = ct.valkey.Do(ctx, cmd).Error()
	if err != nil {
		return err
	}

	// Implementation for tracking the click event
	return nil
}
