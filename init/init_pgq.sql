CREATE EXTENSION IF NOT EXISTS pgq;

select * from pgq.create_queue('error_queue');

select * from pgq.register_consumer('error_queue', 'telegram_bot_consumer');