BEGIN;

DO $$ BEGIN
  CREATE TYPE broadcasterType AS ENUM ('partner', 'affiliate', 'normal');
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS vg_options (
  singlerow_id BOOL PRIMARY KEY DEFAULT TRUE,
  last_reconciliation_at timestamp,
  CONSTRAINT singlerow_unique CHECK (singlerow_id)
);

CREATE TABLE IF NOT EXISTS tracked_channels (
  broadcaster_id varchar PRIMARY KEY,
  broadcaster_display_name varchar(25) NOT NULL,
  broadcaster_username varchar(25) NOT NULL,
  broadcaster_type broadcasterType NOT NULL,
  profile_image_url varchar,
  offline_image_url varchar,
  tracked_since timestamp NOT NULL
);

COMMIT;

