ALTER TABLE trust_center_accesses
  ADD COLUMN electronic_signature_id TEXT REFERENCES electronic_signatures(id);