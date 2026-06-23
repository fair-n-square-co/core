-- +goose Up
-- +goose StatementBegin
-- This migration defines the lean, WorkOS-aligned schema for users and friendships.
--
-- Core does not own user identity or profile data: authentication is handled by an
-- external provider (WorkOS) and the canonical user record lives in the Authx service
-- (see ADR-2, ADR-4). Here we keep only our own generated id plus a reference to the
-- external auth subject, so we can leverage database features like ON DELETE CASCADE
-- and own our keyspace independently of the auth provider.
--
-- Friendships are modelled as a "relationship" (current state, one row per pair) plus
-- a "friend_event" append-only history. Direction is captured by status_actor_id (who
-- caused the current state) and friend_event.actor_id (who performed each event).

-- Create a function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- user is a local reference to an identity owned by the external auth provider / Authx.
-- We generate our own id and store the external auth subject separately, so FKs never
-- depend on the provider's keyspace (the provider must be swappable per ADR-4).
CREATE TABLE "user" (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  auth_subject  text NOT NULL UNIQUE,
  created_at    timestamptz NOT NULL DEFAULT now()
);

-- relationship holds the current state of a friendship between a pair of users.
-- The pair is ordered once at creation (user_a < user_b) so each pair maps to exactly
-- one row via a plain UNIQUE constraint -- no LEAST/GREATEST expression index needed.
CREATE TABLE relationship (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_a          uuid NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  user_b          uuid NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  status          text NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'accepted', 'rejected', 'cancelled', 'blocked')),
  -- who caused the current status (e.g. the requester while pending, the blocker while blocked)
  status_actor_id uuid REFERENCES "user"(id) ON DELETE SET NULL,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ordered_users CHECK (user_a < user_b),
  CONSTRAINT unique_pair   UNIQUE (user_a, user_b)
);

-- Indexes for listing a user's relationships filtered by status, in both pair positions.
CREATE INDEX idx_relationship_user_a ON relationship (user_a, status);
CREATE INDEX idx_relationship_user_b ON relationship (user_b, status);

-- Trigger to automatically update updated_at on row update
CREATE TRIGGER update_relationship_updated_at BEFORE UPDATE ON relationship
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- friend_event is the append-only history of lifecycle events for a relationship.
CREATE TABLE friend_event (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  relationship_id uuid NOT NULL REFERENCES relationship(id) ON DELETE CASCADE,
  actor_id        uuid NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  type            text NOT NULL
                    CHECK (type IN ('requested', 'accepted', 'rejected', 'cancelled', 'blocked', 'unblocked')),
  created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_friend_event_relationship ON friend_event (relationship_id, created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_friend_event_relationship;
DROP TABLE IF EXISTS friend_event;

DROP TRIGGER IF EXISTS update_relationship_updated_at ON relationship;
DROP INDEX IF EXISTS idx_relationship_user_b;
DROP INDEX IF EXISTS idx_relationship_user_a;
DROP TABLE IF EXISTS relationship;

DROP TABLE IF EXISTS "user";

DROP FUNCTION IF EXISTS update_updated_at_column();

-- +goose StatementEnd
