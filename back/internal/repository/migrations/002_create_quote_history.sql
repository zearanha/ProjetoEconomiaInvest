CREATE TABLE IF NOT EXISTS quote_history (
    symbol TEXT NOT NULL,
    market TEXT NOT NULL,
    sampled_at TIMESTAMPTZ NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    currency TEXT NOT NULL,
    PRIMARY KEY (symbol, sampled_at)
);

CREATE INDEX IF NOT EXISTS quote_history_sampled_at_idx ON quote_history (sampled_at DESC);
