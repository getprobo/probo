ALTER TABLE iam_oauth2_consents
  ADD COLUMN session_id TEXT NOT NULL REFERENCES iam_sessions(id);
