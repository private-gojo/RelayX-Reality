package config

func serverOutbounds() []any {
	return []any{
		map[string]any{"protocol": "freedom", "tag": "direct"},
	}
}
