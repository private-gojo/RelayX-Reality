package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/private-gojo/relayx-reality/internal/config"
	"github.com/private-gojo/relayx-reality/internal/relay"
	xrayconfig "github.com/private-gojo/relayx-reality/internal/xray/config"
	xrayprocess "github.com/private-gojo/relayx-reality/internal/xray/process"
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)
	if len(os.Args) < 2 {
		fatal("usage: relay server | client | client-config | validate-config")
	}

	var err error
	switch os.Args[1] {
	case "server":
		err = relay.RunServer()
	case "client":
		err = runClient()
	case "client-config":
		err = writeClientConfig()
	case "validate-config":
		err = validateXrayConfig()
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fatal("%v", err)
	}
}

func runClient() error {
	localAddr := valueOr(os.Getenv("LOCAL_ADDR"), "127.0.0.1:11000")
	remoteURL := os.Getenv("REMOTE_WS_URL")
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return relay.RunClient(ctx, cfg, localAddr, remoteURL)
}

func writeClientConfig() error {
	cfg, err := config.LoadClient()
	if err != nil {
		return err
	}
	data, err := xrayconfig.GenerateClientJSON(cfg)
	if err != nil {
		return err
	}
	output := valueOr(os.Getenv("CLIENT_CONFIG_OUT"), "client/xray-client.json")
	if err := os.MkdirAll(filepath.Dir(output), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(output, data, 0600); err != nil {
		return err
	}
	log.Printf("[CLIENT] Xray config written to %s", output)
	return nil
}

func validateXrayConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	serverConfig, err := xrayconfig.GenerateServerJSON(cfg)
	if err != nil {
		return err
	}
	serverPath, cleanupServer, err := xrayprocess.WriteTempData(serverConfig)
	if err != nil {
		return err
	}
	defer cleanupServer()

	clientConfig, err := xrayconfig.GenerateClientJSON(cfg)
	if err != nil {
		return err
	}
	clientPath, cleanupClient, err := xrayprocess.WriteTempData(clientConfig)
	if err != nil {
		return err
	}
	defer cleanupClient()

	xrayBin := valueOr(os.Getenv("XRAY_BIN"), "/usr/local/bin/xray")
	if err := xrayprocess.TestConfig(xrayBin, serverPath); err != nil {
		return fmt.Errorf("server config validation failed: %w", err)
	}
	if err := xrayprocess.TestConfig(xrayBin, clientPath); err != nil {
		return fmt.Errorf("client config validation failed: %w", err)
	}
	log.Printf("[CHECK] server and client Xray configurations accepted")
	return nil
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func fatal(format string, args ...any) {
	log.Printf("[FATAL] "+format, args...)
	os.Exit(1)
}
