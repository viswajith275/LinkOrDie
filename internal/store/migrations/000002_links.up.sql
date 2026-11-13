CREATE TABLE collections (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, name)
);

CREATE TRIGGER update_collections_modtime BEFORE UPDATE ON collections
  FOR EACH ROW EXECUTE FUNCTION update_modified_column();

CREATE TYPE link_state AS ENUM ('pending','alive','flaky','dead');
CREATE TYPE check_outcome AS ENUM (
  'ok','redirect_ok','blocked','not_found','gone','dns_failure',
  'server_error','timeout','conn_refused','tls_error','redirect_loop','other_error'
);

CREATE TABLE links (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  collection_id        UUID REFERENCES collections(id) ON DELETE SET NULL,
  url                  TEXT NOT NULL,
  normalized_url       TEXT NOT NULL,
  url_hash             BYTEA NOT NULL,
  host                 TEXT NOT NULL,
  title                TEXT,
  state                link_state NOT NULL DEFAULT 'pending',
  last_outcome         check_outcome,
  last_status_code     INT,
  last_error           TEXT,
  cause_of_death       check_outcome,
  consecutive_failures INT NOT NULL DEFAULT 0,
  total_failures       INT NOT NULL DEFAULT 0,
  resurrected_count    INT NOT NULL DEFAULT 0,
  last_checked_at      TIMESTAMPTZ,
  last_ok_at           TIMESTAMPTZ,
  died_at              TIMESTAMPTZ,
  next_check_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, url_hash)
);

CREATE INDEX links_user_idx ON links (user_id, collection_id, state);
CREATE INDEX links_due_idx  ON links (next_check_at);
CREATE INDEX links_user_state_died_idx  ON links (user_id, state, died_at DESC NULLS LAST);
CREATE TRIGGER update_links_modtime BEFORE UPDATE ON links
  FOR EACH ROW EXECUTE FUNCTION update_modified_column();

CREATE TABLE checks (
  id          BIGSERIAL PRIMARY KEY,
  link_id     UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
  checked_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  outcome     check_outcome NOT NULL,
  status_code INT,
  final_url   TEXT,
  method      TEXT NOT NULL,
  elapsed_ms  INT NOT NULL,
  error       TEXT
);
CREATE INDEX checks_link_idx ON checks (link_id, checked_at DESC);
CREATE INDEX checks_checked_at_idx ON checks (checked_at DESC);

CREATE TABLE transitions (
  id          BIGSERIAL PRIMARY KEY,
  link_id     UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
  from_state  link_state NOT NULL,
  to_state    link_state NOT NULL,
  outcome     check_outcome,
  cause       check_outcome,
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  acknowledged_at TIMESTAMPTZ

);
CREATE INDEX transitions_link_idx ON transitions (link_id, occurred_at DESC);
CREATE INDEX transitions_unacked_idx ON transitions (occurred_at DESC) WHERE to_state = 'alive' AND acknowledged_at IS NULL;
