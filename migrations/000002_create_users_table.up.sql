CREATE TABLE users (
    id            UUID PRIMARY KEY,
    username      TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    email         TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_email_unique ON users (LOWER(email));

CREATE UNIQUE INDEX users_username_unique ON users (LOWER(username));