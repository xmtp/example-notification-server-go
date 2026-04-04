package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/xmtp/example-notification-server-go/pkg/api"
	database "github.com/xmtp/example-notification-server-go/pkg/db"
	"github.com/xmtp/example-notification-server-go/pkg/delivery"
	"github.com/xmtp/example-notification-server-go/pkg/installations"
	"github.com/xmtp/example-notification-server-go/pkg/interfaces"
	"github.com/xmtp/example-notification-server-go/pkg/logging"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"github.com/xmtp/example-notification-server-go/pkg/subscriptions"
	"github.com/xmtp/example-notification-server-go/pkg/xmtp"
	"go.uber.org/zap"
)

var opts options.Options
var logger *zap.Logger

var (
	GitCommit           string
	XMTPGoClientVersion string
)

func main() {
	var err error
	if _, err = flags.Parse(&opts); err != nil {
		if err, ok := err.(*flags.Error); !ok || err.Type != flags.ErrHelp {
			log.Fatalf("Could not parse options: %s", err)
		}
		return
	}

	logger = logging.CreateLogger(opts.LogEncoding, opts.LogLevel)

	clientVersion := "example-notifications-server-go/" + shortGitCommit()
	appVersion := "xmtp-go/" + shortXMTPGoClientVersion()

	logger.Info("starting", zap.String("client-version", clientVersion), zap.String("app-version", appVersion))

	if opts.CreateMigration != "" {
		if err = createMigration(); err != nil {
			logger.Fatal("failed to create migration", zap.Error(err))
		}
		return
	}

	if !opts.Xmtp.ListenerEnabled && !opts.Api.Enabled {
		logger.Fatal("no --api or --xmtp-listener flags applied")
	}

	db := initDb()
	ctx, cancel := context.WithCancel(context.Background())
	installationsService := installations.NewInstallationsService(logger, db)
	subscriptionsService := subscriptions.NewSubscriptionsService(logger, db)
	var notifListener xmtp.NotificationListener
	var apiServer *api.ApiServer

	if opts.Xmtp.ListenerEnabled {
		deliveryServices := []interfaces.Delivery{}
		var err error

		if opts.Apns.Enabled {
			apns, err := delivery.NewApnsDelivery(logger, opts.Apns)
			if err != nil {
				logger.Fatal("failed to initialize APNS", zap.Error(err))
			}
			deliveryServices = append(deliveryServices, apns)
		}

		if opts.Fcm.Enabled {
			fcm, err := delivery.NewFcmDelivery(ctx, logger, opts.Fcm)
			if err != nil {
				logger.Fatal("failed to initialize FCM", zap.Error(err))
			}
			deliveryServices = append(deliveryServices, fcm)
		}

		if opts.HttpDelivery.Enabled {
			deliveryServices = append(deliveryServices, delivery.NewHttpDelivery(logger, opts.HttpDelivery))
		}

		switch opts.Xmtp.ListenerType {
		case "v4":
			notifListener, err = xmtp.NewV4Listener(ctx, logger, opts.Xmtp, installationsService, subscriptionsService, deliveryServices, clientVersion, appVersion)
		default: // "v3"
			notifListener, err = xmtp.NewListener(ctx, logger, opts.Xmtp, installationsService, subscriptionsService, deliveryServices, clientVersion, appVersion)
		}
		if err != nil {
			logger.Fatal("failed to initialize listener", zap.Error(err))
		}
		notifListener.Start()
	}

	if opts.Api.Enabled {
		apiServer = api.NewApiServer(logger, opts.Api, installationsService, subscriptionsService, interfaces.ListenerType(opts.Xmtp.ListenerType))
		apiServer.Start()
	}

	waitForShutdown()

	if apiServer != nil {
		apiServer.Stop()
	}

	if notifListener != nil {
		notifListener.Stop()
	}

	cancel()
}

// Commenting out as these are currently unused
func waitForShutdown() {
	termChannel := make(chan os.Signal, 1)
	signal.Notify(termChannel, syscall.SIGINT, syscall.SIGTERM)
	<-termChannel
}

func initDb() *sql.DB {
	db, err := database.CreateDB(opts.DbConnectionString, 10*time.Second)
	if err != nil {
		log.Fatal("db creation error", zap.Error(err))
	}

	err = database.Migrate(context.Background(), db)
	if err != nil {
		log.Fatal("db migration error", zap.Error(err))
	}

	return db
}

func createMigration() error {
	files, err := database.CreateMigrationFiles(opts.CreateMigration)
	if err != nil {
		return err
	}
	for _, file := range files {
		fmt.Printf("created migration %s (%s)\n", file.Name, file.Path)
	}
	return nil
}

func shortGitCommit() string {
	val := GitCommit
	if len(val) >= 7 {
		val = val[:7]
	}
	return val
}

func shortXMTPGoClientVersion() string {
	return strings.Split(XMTPGoClientVersion, "-")[0]
}
