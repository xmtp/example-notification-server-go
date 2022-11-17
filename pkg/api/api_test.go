package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/xmtp/example-notification-server-go/pkg/proto/protoconnect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func buildClient() protoconnect.NotificationsClient {
	return protoconnect.NewNotificationsClient(
		http.DefaultClient,
		"http://localhost:8080",
	)
}

func startServer(t *testing.T, server ApiServer) func() {
	ctx := context.Background()
	mux := http.NewServeMux()
	path, handler := protoconnect.NewNotificationsHandler(&server)
	mux.Handle(path, handler)
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {

		}
	}()

	return func() {
		httpServer.Shutdown(ctx)
	}
}

// func mockInstallations
