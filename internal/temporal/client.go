package temporal

import (
	"github.com/robryanx/mcp-temporal-server/internal/config"
	"go.temporal.io/sdk/client"
)

func NewTemporalClient(cfg config.Config) (client.Client, error) {
	return client.NewClient(client.Options{
		HostPort:  cfg.TemporalAddress,
		Namespace: cfg.Namespace,
	})
}
