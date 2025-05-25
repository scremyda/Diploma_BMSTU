CREATE EXTENSION IF NOT EXISTS pgmq CASCADE;

CREATE TABLE IF NOT EXISTS certificates_status (
   id            BIGSERIAL       PRIMARY KEY,
   msg_id        BIGINT          NOT NULL UNIQUE,
   payload       JSONB           NOT NULL,
   processed_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);