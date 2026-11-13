package config

import (
	"encoding/json"

	appconfig "github.com/private-gojo/relayx-reality/internal/config"
)

func GenerateServerJSON(c appconfig.Config) ([]byte, error) {
	serverConfig := map[string]any{
		"log":       map[string]any{"loglevel": "warning"},
		"inbounds":  serverInbounds(c),
		"outbounds": serverOutbounds(),
	}
	return json.MarshalIndent(serverConfig, "", "  ")
}
