CREATE EXTENSION IF NOT EXISTS pgq;

DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1 FROM pg_catalog.pg_class
            WHERE relname = 'error_queue'
        ) THEN
            PERFORM pgq.create_queue('error_queue');
        END IF;
    END
$$;
