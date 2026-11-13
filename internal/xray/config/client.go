package config

import (
	"encoding/json"

	appconfig "github.com/private-gojo/relayx-reality/internal/config"
)

func GenerateClientJSON(c appconfig.Config) ([]byte, error) {
	publicKey, err := c.PublicKey()
	if err != nil {
		return nil, err
	}
	clientConfig := map[string]any{
		"log": map[string]any{"loglevel": "warning"},
		"inbounds": []any{
			map[string]any{
				"listen":   "127.0.0.1",
				"port":     10808,
				"protocol": "socks",
				"settings": map[string]any{"udp": true},
			},
		},
		"outbounds": []any{
			map[string]any{
				"tag":      "proxy",
				"protocol": "vless",
				"settings": map[string]any{
					"vnext": []any{
						map[string]any{
							"address": "127.0.0.1",
							"port":    11000,
							"users": []any{
								map[string]any{"id": c.UUID, "encryption": "none"},
							},
						},
					},
				},
				"streamSettings": map[string]any{
					"method":   "xhttp",
					"security": "reality",
					"realitySettings": map[string]any{
						"serverName":  c.ServerName,
						"fingerprint": "chrome",
						"password":    publicKey,
						"shortId":     c.ShortID,
					},
					"xhttpSettings": map[string]any{
						"path": c.Path,
						"mode": "auto",
					},
				},
			},
		},
	}
	return json.MarshalIndent(clientConfig, "", "  ")
}
