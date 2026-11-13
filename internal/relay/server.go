package relay

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/private-gojo/relayx-reality/internal/bridge/server"
	"github.com/private-gojo/relayx-reality/internal/config"
	xrayconfig "github.com/private-gojo/relayx-reality/internal/xray/config"
	xrayprocess "github.com/private-gojo/relayx-reality/internal/xray/process"
)

func RunServer() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	serverConfig, err := xrayconfig.GenerateServerJSON(cfg)
	if err != nil {
		return err
	}
	configPath, cleanup, err := xrayprocess.WriteTempData(serverConfig)
	if err != nil {
		return err
	}
	defer cleanup()

	xrayBin := valueOr(os.Getenv("XRAY_BIN"), "/usr/local/bin/xray")
	if err := xrayprocess.TestConfig(xrayBin, configPath); err != nil {
		return err
	}

	xrayCmd := exec.Command(xrayBin, "run", "-config", configPath)
	xrayCmd.Stdout = os.Stdout
	xrayCmd.Stderr = os.Stderr
	if err := xrayCmd.Start(); err != nil {
		return fmt.Errorf("start Xray: %w", err)
	}
	xrayDone := make(chan error, 1)
	go func() { xrayDone <- xrayCmd.Wait() }()

	if err := xrayprocess.WaitForTCP("127.0.0.1:10000", 10*time.Second); err != nil {
		_ = xrayCmd.Process.Kill()
		<-xrayDone
		return fmt.Errorf("Xray inbound did not become ready: %w", err)
	}

	bridgeServer := server.NewServer("127.0.0.1:10000", cfg.Path, cfg.BridgeToken, cfg.MaxSessions)
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           bridgeServer.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	httpDone := make(chan error, 1)
	go func() {
		log.Printf("[SERVER] listening=:%s path=%s max_sessions=%d", cfg.Port, cfg.Path, cfg.MaxSessions)
		httpDone <- httpServer.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	var runErr error
	select {
	case sig := <-signals:
		log.Printf("[SERVER] received signal=%s", sig)
	case err := <-httpDone:
		if !errors.Is(err, http.ErrServerClosed) {
			runErr = fmt.Errorf("HTTP server stopped: %w", err)
			log.Printf("[SERVER] %v", runErr)
		}
	case err := <-xrayDone:
		bridgeServer.CloseAll()
		_ = httpServer.Close()
		return fmt.Errorf("Xray exited unexpectedly: %v", err)
	}

	bridgeServer.CloseAll()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	_ = xrayCmd.Process.Signal(os.Interrupt)
	select {
	case <-xrayDone:
	case <-time.After(5 * time.Second):
		_ = xrayCmd.Process.Kill()
		<-xrayDone
	}
	return runErr
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
