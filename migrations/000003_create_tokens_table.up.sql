CREATE TABLE tokens (
    id            SERIAL PRIMARY KEY,
    user_id       UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    refresh_token TEXT NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL
);