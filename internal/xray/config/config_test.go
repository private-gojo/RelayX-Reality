package config

import (
	"encoding/json"
	"testing"

	appconfig "github.com/private-gojo/relayx-reality/internal/config"
)

func validXrayConfig() appconfig.Config {
	return appconfig.Config{
		UUID:        "47be59a9-103b-4065-9c3a-ebb9ecf8ffd3",
		PrivateKey:  appconfig.FixedPrivateKey,
		ShortID:     "0123456789abcdef",
		ServerName:  appconfig.FixedServerName,
		Target:      appconfig.FixedTarget,
		Path:        "/reality",
		BridgeToken: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Port:        "8080",
		MaxSessions: 256,
	}
}

func TestServerJSONUsesFixedFirebaseREALITYDestination(t *testing.T) {
	data, err := GenerateServerJSON(validXrayConfig())
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	inbound := decoded["inbounds"].([]any)[0].(map[string]any)
	stream := inbound["streamSettings"].(map[string]any)
	if stream["method"] != "xhttp" || stream["security"] != "reality" {
		t.Fatalf("unexpected transport settings: %#v", stream)
	}
	reality := stream["realitySettings"].(map[string]any)
	if reality["target"] != appconfig.FixedTarget {
		t.Fatalf("unexpected REALITY target: %v", reality["target"])
	}
	serverNames := reality["serverNames"].([]any)
	if len(serverNames) != 1 || serverNames[0] != appconfig.FixedServerName {
		t.Fatalf("unexpected REALITY serverNames: %#v", serverNames)
	}
	if _, exists := stream["network"]; exists {
		t.Fatal("legacy network field must not be emitted")
	}
}

func TestClientJSONUsesFixedFirebaseREALITYServerName(t *testing.T) {
	data, err := GenerateClientJSON(validXrayConfig())
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	outbound := decoded["outbounds"].([]any)[0].(map[string]any)
	stream := outbound["streamSettings"].(map[string]any)
	reality := stream["realitySettings"].(map[string]any)
	if reality["password"] != appconfig.FixedPublicKey {
		t.Fatalf("unexpected REALITY password/public key: %v", reality["password"])
	}
	if reality["serverName"] != appconfig.FixedServerName {
		t.Fatalf("unexpected REALITY serverName: %v", reality["serverName"])
	}
	if _, exists := reality["publicKey"]; exists {
		t.Fatal("legacy publicKey field must not be emitted")
	}
}
