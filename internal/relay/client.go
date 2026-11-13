package relay

import (
	"context"

	"github.com/private-gojo/relayx-reality/internal/bridge/client"
	"github.com/private-gojo/relayx-reality/internal/config"
)

func RunClient(ctx context.Context, cfg config.Config, localAddr, remoteURL string) error {
	bridgeClient, err := client.NewClient(localAddr, remoteURL, cfg.BridgeToken, cfg.MaxSessions)
	if err != nil {
		return err
	}
	return bridgeClient.Run(ctx)
}
