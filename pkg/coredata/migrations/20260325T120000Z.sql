CREATE TYPE search_engine_indexing AS ENUM ('INDEXABLE', 'NOT_INDEXABLE');

UPDATE trust_centers
SET search_engine_indexing = 'NOT_INDEXABLE'
WHERE search_engine_indexing = '';

ALTER TABLE trust_centers
ALTER COLUMN search_engine_indexing TYPE search_engine_indexing
USING search_engine_indexing::search_engine_indexing;
