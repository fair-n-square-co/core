-- +goose Up
-- +goose StatementBegin
-- This migration is focused on defining a friendship between personas.
-- Friends for a persona

-- Create a function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TYPE friend_status AS ENUM ('pending', 'accepted', 'blocked');

-- Persona table is a record for a persona.
-- While the persona is maintained by external system,
-- we keep a very basic copy in a table so we can leverage database features like
-- ON DELETE CASCADE if external system deletes this persona
CREATE TABLE persona (
  -- id is the persona id fetched from JWT/API request and not auto generated.
  id uuid PRIMARY KEY NOT NULL,
  created_at timestamp with time zone DEFAULT now()
);

CREATE TABLE friend (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  persona_id_1 uuid NOT NULL REFERENCES persona(id),
  persona_id_2 uuid NOT NULL REFERENCES persona(id),
  status friend_status NOT NULL DEFAULT 'pending',
  created_at timestamp with time zone DEFAULT now(),
  updated_at timestamp with time zone DEFAULT now(),
  -- persona2 can record a friendly name for persona 1 and vice versa
  name_persona_1 text NOT NULL,
  name_persona_2 text NOT NULL,

  CONSTRAINT different_personas CHECK (persona_id_1 <> persona_id_2)
);

-- Index for queries to filter friends in both directions
CREATE INDEX idx_friendship_1 ON friend (status, persona_id_1);
CREATE INDEX idx_friendship_2 ON friend (status, persona_id_2);

-- Prevent duplicate friend requests in both directions
-- Use LEAST/GREATEST to normalize persona pairs so (A,B) and (B,A)
-- are treated as the same friendship
-- People can block each other and unblock each other,
-- so there can be two blocks in that scenario
CREATE UNIQUE INDEX unique_friendship ON friend
(
  LEAST(persona_id_1, persona_id_2),
  GREATEST(persona_id_1, persona_id_2)
) WHERE status IN ('pending', 'accepted');

-- Trigger to automatically update updated_at on row update
CREATE TRIGGER update_friend_updated_at BEFORE UPDATE ON friend
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS update_friend_updated_at ON friend;

DROP INDEX IF EXISTS unique_friendship;
DROP INDEX IF EXISTS idx_friendship_2;
DROP INDEX IF EXISTS idx_friendship_1;

DROP TABLE IF EXISTS "friend" CASCADE;
DROP TABLE IF EXISTS "persona" CASCADE;
DROP TYPE friend_status;

DROP FUNCTION IF EXISTS update_updated_at_column();

-- +goose StatementEnd
