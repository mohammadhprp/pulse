CREATE DATABASE IF NOT EXISTS gologcentral;

CREATE TABLE IF NOT EXISTS gologcentral.logs (
    EventTimeMs UInt64,
    Timestamp   DateTime MATERIALIZED toDateTime(EventTimeMs / 1000),
    Service     String,
    Level       Enum8('DEBUG'=1, 'INFO'=2, 'WARN'=3, 'ERROR'=4),
    Message     String,
    Host        String,
    RequestID   UUID
) ENGINE = MergeTree
PARTITION BY toYYYYMMDD(Timestamp)
ORDER BY (Service, Level, Timestamp)
TTL Timestamp + INTERVAL 30 DAY;