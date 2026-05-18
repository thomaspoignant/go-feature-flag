-- Optional dev seed. Idempotent. Runs after migrations.
-- Note: real users are created on first OIDC login; this only creates
-- a sample team so that you can immediately see something in the UI/API.

INSERT INTO teams (name, description)
VALUES ('demo', 'Demo team created by bootstrap.sh')
ON CONFLICT (name) DO NOTHING;
