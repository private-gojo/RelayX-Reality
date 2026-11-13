package config

import appconfig "github.com/private-gojo/relayx-reality/internal/config"

func serverInbounds(c appconfig.Config) []any {
	return []any{
		map[string]any{
			"listen":   "127.0.0.1",
			"port":     10000,
			"protocol": "vless",
			"settings": map[string]any{
				"clients":    []any{map[string]any{"id": c.UUID}},
				"decryption": "none",
			},
			"streamSettings": map[string]any{
				"method":   "xhttp",
				"security": "reality",
				"realitySettings": map[string]any{
					"show":        false,
					"target":      c.Target,
					"xver":        0,
					"serverNames": []string{c.ServerName},
					"privateKey":  c.PrivateKey,
					"shortIds":    []string{c.ShortID},
				},
				"xhttpSettings": map[string]any{
					"path": c.Path,
					"mode": "auto",
				},
			},
		},
	}
}
