-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users
(
    id         UUID         NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    username   TEXT         NOT NULL UNIQUE,
    password   VARCHAR(100) NOT NULL,
    coins      INT          NOT NULL DEFAULT 100000 CHECK (coins >= 0),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS merch_items
(
    id         UUID        NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    name       VARCHAR(50) NOT NULL UNIQUE,
    price      INT         NOT NULL CHECK (price > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS purchases
(
    id                UUID PRIMARY KEY     DEFAULT uuid_generate_v4(),
    user_id           UUID        NOT NULL,
    merch_id          UUID        NOT NULL,
    price_at_purchase INT         NOT NULL CHECK (price_at_purchase >= 0),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT purchases_user_fk
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT purchases_merch_fk
        FOREIGN KEY (merch_id) REFERENCES merch_items (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS coin_transactions
(
    id           UUID PRIMARY KEY     DEFAULT uuid_generate_v4(),
    from_user_id UUID        NOT NULL,
    to_user_id   UUID        NOT NULL,
    amount       INT         NOT NULL CHECK (amount > 0),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT coin_trans_from_fk
        FOREIGN KEY (from_user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT coin_trans_to_fk
        FOREIGN KEY (to_user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_purchases_user_id ON purchases(user_id);
CREATE INDEX IF NOT EXISTS idx_purchases_merch_id ON purchases(merch_id);

CREATE INDEX IF NOT EXISTS idx_coin_trans_from_user ON coin_transactions(from_user_id);
CREATE INDEX IF NOT EXISTS idx_coin_trans_to_user ON coin_transactions(to_user_id);

INSERT INTO merch_items (name, price)
VALUES
    ('t-shirt', 8000),
    ('cup', 2000),
    ('book', 5000),
    ('pen', 1000),
    ('powerbank', 20000),
    ('hoody', 30000),
    ('umbrella', 20000),
    ('socks', 1000),
    ('wallet', 5000),
    ('pink-hoody', 50000)
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS coin_transactions;
DROP TABLE IF EXISTS purchases;
DROP TABLE IF EXISTS merch_items;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
