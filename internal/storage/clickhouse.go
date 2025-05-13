package storage

import (
	"context"
	"fmt"
	"strings"
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

// QueryEvents retrieves events from ClickHouse with filtering and sorting options
func QueryEvents(ctx context.Context, conn clickhouse.Conn, options models.QueryOptions) ([]models.Event, error) {
	var events []models.Event

	var conditions []string
	var params []interface{}

	if options.Service != "" {
		conditions = append(conditions, "Service = ?")
		params = append(params, options.Service)
	}

	if options.Level != "" {
		conditions = append(conditions, "Level = ?")
		params = append(params, options.Level)
	}

	if options.Host != "" {
		conditions = append(conditions, "Host = ?")
		params = append(params, options.Host)
	}

	if options.StartTime > 0 {
		conditions = append(conditions, "EventTimeMs >= ?")
		params = append(params, options.StartTime)
	}

	if options.EndTime > 0 {
		conditions = append(conditions, "EventTimeMs <= ?")
		params = append(params, options.EndTime)
	}

	if options.RequestID != "" {
		conditions = append(conditions, "RequestID = ?")
		params = append(params, options.RequestID)
	}

	if options.SearchQuery != "" {
		conditions = append(conditions, "Message LIKE ?")
		params = append(params, "%"+options.SearchQuery+"%")
	}

	if options.Limit <= 0 {
		options.Limit = 100
	}

	sortOrder := "ASC"
	if strings.ToUpper(options.SortOrder) == "DESC" {
		sortOrder = "DESC"
	}

	query := "SELECT EventTimeMs, Service, Level, Message, Host, RequestID FROM gologcentral.logs"

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY EventTimeMs %s", sortOrder)

	query += fmt.Sprintf(" LIMIT %d", options.Limit)
	if options.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", options.Offset)
	}

	logger.Debug("Executing query",
		zap.String("query", query),
		zap.Any("params", params))

	start := time.Now()

	rows, err := conn.Query(ctx, query, params...)
	if err != nil {
		logger.Error("Failed to query events", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event models.Event
		err := rows.Scan(&event.EventTimeMs, &event.Service, &event.Level, &event.Message, &event.Host, &event.RequestID)
		if err != nil {
			logger.Error("Failed to scan row", zap.Error(err))
			return events, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error during row iteration", zap.Error(err))
		return events, err
	}

	logger.Debug("Query completed successfully",
		zap.Duration("took", time.Since(start)),
		zap.Int("count", len(events)))

	return events, nil
}
