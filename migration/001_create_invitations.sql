-- Migration: 001_create_invitations.sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS invitations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL,
    slug         VARCHAR(60) NOT NULL UNIQUE,
    status       VARCHAR(20) NOT NULL DEFAULT 'draft',
    config       JSONB NOT NULL DEFAULT '{}',
    content      JSONB NOT NULL DEFAULT '{}',
    rsvp_count   INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_invitations_user_id    ON invitations(user_id);
CREATE INDEX idx_invitations_slug       ON invitations(slug);
CREATE INDEX idx_invitations_status     ON invitations(status);
CREATE INDEX idx_invitations_published  ON invitations(published_at) WHERE published_at IS NOT NULL;
