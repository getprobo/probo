-- TODO: drop the classification column from documents.
ALTER TABLE documents ALTER COLUMN classification SET DEFAULT 'SECRET';
