package storage

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/mohammadhptp/pulse/pkg/logger"
	"github.com/mohammadhptp/pulse/pkg/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Connect establishes a connection to ClickHouse
func Connect(ctx context.Context) (clickhouse.Conn, error) {
	addr := viper.GetString("CLICKHOUSE_ADDR")
	db := viper.GetString("CLICKHOUSE_DB")
	user := viper.GetString("CLICKHOUSE_USER")
	pass := viper.GetString("CLICKHOUSE_PASS")

	logger.Info("Connecting to ClickHouse",
		zap.String("address", addr),
		zap.String("database", db))

	return clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: db,
			Username: user,
			Password: pass,
		},
		DialTimeout:     5 * time.Second,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 10 * time.Minute,
	})
}

// InsertEvent inserts a single event into ClickHouse
func InsertEvent(ctx context.Context, conn clickhouse.Conn, e models.Event) error {
	query := "INSERT INTO gologcentral.logs (EventTimeMs, Service, Level, Message, Host, RequestID)"

	batch, err := conn.PrepareBatch(ctx, query)
	if err != nil {
		logger.Error("Failed to prepare batch", zap.Error(err))
		return err
	}

	if err := batch.Append(e.EventTimeMs, e.Service, e.Level, e.Message, e.Host, e.RequestID); err != nil {
		logger.Error("Failed to append to batch",
			zap.Error(err),
			zap.String("service", e.Service),
			zap.Uint64("timestamp", e.EventTimeMs))
		return err
	}

	start := time.Now()
	if err := batch.Send(); err != nil {
		logger.Error("Failed to send batch", zap.Error(err))
		return err
	}

	logger.Debug("Event inserted successfully",
		zap.Duration("took", time.Since(start)),
		zap.String("service", e.Service),
		zap.String("level", e.Level))

	return nil
}
