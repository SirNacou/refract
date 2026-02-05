package snowflake

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

var (
	snowflakeEpoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	Generator      *SnowflakeNode
)

type SnowflakeNode struct {
	*snowflake.Node
}

func NewSnowflakeNode(nodeID int64) error {
	snowflake.Epoch = snowflakeEpoch.UnixMilli()
	node, err := snowflake.NewNode(nodeID)

	if err != nil {
		return err
	}

	Generator = &SnowflakeNode{node}

	return nil
}
