BEGIN;

DO $$ BEGIN
  CREATE TYPE broadcasterType AS ENUM ('partner', 'affiliate', 'normal');
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS tracker_options (
  singlerow_id BOOL PRIMARY KEY DEFAULT TRUE,
  client_secret varchar,
  client_id varchar,
  access_token varchar,
  last_reconciliation_at timestamp,
  CONSTRAINT singlerow_unique CHECK (singlerow_id)
);

CREATE TABLE IF NOT EXISTS tracked_channels (
  tracked_id serial PRIMARY KEY,
  broadcaster_username varchar(25) NOT NULL,
  broadcaster_type broadcasterType NOT NULL,
  profile_image_url varchar,
  offline_image_url varchar,
  tracked_since timestamp NOT NULL
);

COMMIT;

