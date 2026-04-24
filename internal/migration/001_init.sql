-- Migration: 001_init.sql
-- Run once against your PostgreSQL database

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id               TEXT        PRIMARY KEY,
    name             TEXT        NOT NULL UNIQUE,
    phone            TEXT        NOT NULL,
    city             TEXT        NOT NULL,
    skills           TEXT[]      NOT NULL DEFAULT '{}',
    availability     TEXT        NOT NULL DEFAULT '',
    telegram_chat_id BIGINT,
    rating           FLOAT       NOT NULL DEFAULT 0,
    tg_verified      BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_name_lower ON users (LOWER(name));

-- OTP codes (one per user, replaced on each request)
CREATE TABLE IF NOT EXISTS otp_codes (
    user_id    TEXT        PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    code       TEXT        NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

-- Pending registrations (token issued before Telegram activation)
CREATE TABLE IF NOT EXISTS pending_registrations (
    user_id    TEXT        PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pending_token ON pending_registrations (token);
