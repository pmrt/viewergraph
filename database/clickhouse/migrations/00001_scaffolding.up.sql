CREATE TABLE IF NOT EXISTS raw_events (
  ts Datetime,
  username String,
  channel LowCardinality(String),
  event_type Enum8('ban', 'subscription', 'view')
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (username, ts, channel);

-- Event tables contains reconciliated events. These events contain state based
-- on previous neighbor events, e.g.: referrer or has_been_banned. These events
-- cannot change anymore, so only insert to this table events that we are sure
-- that won't change (if a event is recent it could be part of the referrer or
-- other states of future events) or with a 'expiration' mutability, e.g.:
-- after 3 days we could consider that events will never be part of future
-- events.
CREATE TABLE IF NOT EXISTS events (
  ts Datetime,
  username String,
  channel LowCardinality(String),
  referrer LowCardinality(String)
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (channel, ts, referrer);

-- From which channels do users come to the given channel
CREATE TABLE aggregated_flows_by_dst (
  ts Datetime,
  channel LowCardinality(String),
  referrer LowCardinality(String),
  total_users AggregateFunction(uniq, String)
) ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (channel, ts, referrer);

CREATE MATERIALIZED VIEW aggregated_flows_by_dst_mv
TO aggregated_flows_by_dst
AS
  SELECT
    ts,
    channel,
    referrer,
    uniqState(username) as total_users
  FROM events
  GROUP BY channel, ts, referrer
  ORDER BY (channel, ts, referrer);

-- To which channels do users go from the given channel
CREATE TABLE aggregated_flows_by_src (
  ts Datetime,
  channel LowCardinality(String),
  referrer LowCardinality(String),
  total_users AggregateFunction(uniq, String)
) ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (referrer, ts, channel);


CREATE MATERIALIZED VIEW aggregated_flows_by_src_mv
TO aggregated_flows_by_src
AS
  SELECT
    ts, channel, referrer,
    uniqMerge(total_users)
  FROM aggregated_flows_by_dst
  GROUP BY referrer, ts, channel
  ORDER BY (referrer, ts, channel);
