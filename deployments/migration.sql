CREATE TABLE IF NOT EXISTS subscriptions (
    id CHAR(40) PRIMARY KEY,
    service VARCHAR(255) NOT NULL,
    cost INTEGER NOT NULL CHECK (cost >= 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NULL,
    CONSTRAINT start_before_end CHECK (end_date IS NULL OR start_date <= end_date)
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_subs_service_user_start ON subscriptions(service, user_id, start_date);
CREATE INDEX IF NOT EXISTS ix_subs_period ON subscriptions(start_date, end_date);