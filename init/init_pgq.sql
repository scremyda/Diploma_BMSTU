CREATE EXTENSION IF NOT EXISTS pgq;

-- select * from pgq.create_queue('error_queue');
--
-- select * from pgq.register_consumer('error_queue', 'telegram_bot_consumer');

-- select * from pgq.next_batch('error_queue', 'telegram_bot_consumer');
--
-- select * from pgq.current_event_table('error_queue');


-- SELECT pgq.insert_event(
--                'error_queue',
--                'error',
--                '{"target": "https://example.com", "message": "Test error message"}'
--        );
