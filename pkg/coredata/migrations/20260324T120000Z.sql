ALTER TABLE trust_centers
ADD COLUMN search_engine_indexing TEXT NOT NULL DEFAULT 'NOT_INDEXABLE';

ALTER TABLE trust_centers
ALTER COLUMN search_engine_indexing DROP DEFAULT;
