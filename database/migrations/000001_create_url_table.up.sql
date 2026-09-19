CREATE TABLE url (
    id            uuid        PRIMARY KEY,
    short_url     varchar(10) NOT NULL UNIQUE,
    long_url      TEXT        NOT NULL,
    long_url_hash bytea       NOT NULL UNIQUE,
    created_at    timestamptz NOT NULL DEFAULT now()
);
