CREATE TABLE IF NOT EXISTS users (
    id          UUID      PRIMARY KEY,
    login       TEXT      UNIQUE NOT NULL,
    password_hash TEXT    NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tokens (
    id            SERIAL    PRIMARY KEY,
    refresh_token VARCHAR(64) NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS users_tokens (
    user_id  UUID REFERENCES users (id) ON DELETE CASCADE,
    token_id INT  REFERENCES tokens (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, token_id) 
);

CREATE TABLE IF NOT EXISTS messages (
    id         SERIAL PRIMARY KEY,
    sender_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);