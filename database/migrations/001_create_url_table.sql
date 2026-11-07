CREATE TABLE IF NOT EXISTS url (
    id            UUID PRIMARY KEY,
    long_url      TEXT        NOT NULL,
    long_url_hash BYTEA       NOT NULL UNIQUE, -- dedup key; long_url is too large to index
    short_url     TEXT        NOT NULL UNIQUE, -- constraint name url_short_url_key is relied on by the repository
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
