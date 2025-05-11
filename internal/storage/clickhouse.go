package storage

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/mohammadhptp/pulse/pkg/models"
	"github.com/spf13/viper"
)

func Connect(ctx context.Context) (clickhouse.Conn, error) {
	return clickhouse.Open(&clickhouse.Options{
		Addr: []string{viper.GetString("CLICKHOUSE_ADDR")},
		Auth: clickhouse.Auth{
			Database: viper.GetString("CLICKHOUSE_DB"),
			Username: viper.GetString("CLICKHOUSE_USER"),
			Password: viper.GetString("CLICKHOUSE_PASS"),
		},
		DialTimeout: 5 * time.Second,
	})
}

func InsertEvent(ctx context.Context, conn clickhouse.Conn, e models.Event) error {
	batch, err := conn.PrepareBatch(ctx, "INSERT INTO gologcentral.logs (EventTimeMs, Service, Level, Message, Host, RequestID)")
	if err != nil {
		return err
	}
	if err := batch.Append(e.EventTimeMs, e.Service, e.Level, e.Message, e.Host, e.RequestID); err != nil {
		return err
	}
	return batch.Send()
}
