CREATE TABLE IF NOT EXISTS users (
    id          TEXT PRIMARY KEY,       -- UUID
    email       TEXT UNIQUE NOT NULL,
    password    TEXT NOT NULL,          -- bcrypt hash
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS clients (
    id            TEXT PRIMARY KEY,     -- client_id
    secret        TEXT NOT NULL,        -- bcrypt-hashed client_secret
    name          TEXT NOT NULL,
    redirect_uri  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    token       TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id),
    expires_at  DATETIME NOT NULL,
    revoked     INTEGER DEFAULT 0       -- SQLite has no BOOL, use 0/1
);

CREATE TABLE IF NOT EXISTS authorization_codes (
    code        TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id),
    client_id   TEXT NOT NULL REFERENCES clients(id),
    expires_at  DATETIME NOT NULL,
    used        INTEGER DEFAULT 0
);