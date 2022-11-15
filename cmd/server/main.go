package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/uptrace/bun/migrate"
	"github.com/xmtp/example-notification-server-go/pkg/db"
	"github.com/xmtp/example-notification-server-go/pkg/db/migrations"
	"github.com/xmtp/example-notification-server-go/pkg/logging"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"go.uber.org/zap"
)

var opts options.Options

func main() {
	var err error
	if _, err = flags.Parse(&opts); err != nil {
		if err, ok := err.(*flags.Error); !ok || err.Type != flags.ErrHelp {
			log.Fatalf("Could not parse options: %s", err)
		}
		return
	}

	logger := logging.CreateLogger(opts.LogEncoding, opts.LogLevel)

	if opts.CreateMigration != "" {
		if err = createMigration(); err != nil {
			logger.Fatal("failed to create migration", zap.Error(err))
		}
		return
	}

	// s, err := server.New(logger, installationService, subscriptionService, deliveryService)
	// if err != nil {
	// 	fatal("failed to start server: %s", err)
	// }

	// err = s.RunUntilShutdown()
	// if err != nil {
	// 	logger.Fatal("Failed to start server", zap.Error(err))
	// }

}

func createMigration() error {
	db, err := db.CreateBunDB(opts.DbConnectionString, 30*time.Second)
	if err != nil {
		return err
	}

	migrator := migrate.NewMigrator(db, migrations.Migrations)
	files, err := migrator.CreateSQLMigrations(context.Background(), opts.CreateMigration)
	for _, mf := range files {
		fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
	}

	return err
}
