-- Migration: 002_create_rsvps.sql
CREATE TABLE IF NOT EXISTS rsvps (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invitation_id UUID NOT NULL REFERENCES invitations(id) ON DELETE CASCADE,
    guest_name    VARCHAR(100) NOT NULL,
    status        VARCHAR(20) NOT NULL,
    guest_count   INTEGER NOT NULL DEFAULT 1,
    message       TEXT,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_rsvps_invitation_id ON rsvps(invitation_id);
CREATE INDEX idx_rsvps_status        ON rsvps(status);
CREATE INDEX idx_rsvps_created_at    ON rsvps(created_at);
