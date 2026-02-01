CREATE TABLE users (
    id UUID PRIMARY KEY NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_users_username ON users(username);

CREATE TABLE products (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id),
    url         TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE product_versions (
    id                UUID PRIMARY KEY,
    product_id        UUID NOT NULL REFERENCES products(id),
    version           INTEGER NOT NULL,
    -- when data was read from the source to create a new version
    retrieved_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    name              TEXT NOT NULL,
    description       TEXT,
    price_small_unit  BIGINT NOT NULL,
    currency          TEXT NOT NULL,
    availability      TEXT,
    raw_json        JSONB NOT NULL,

    UNIQUE (product_id, version)
);

CREATE INDEX idx_product_versions_product_id_version_desc
    ON product_versions(product_id, version DESC);
