CREATE TABLE trust_center_document_accesses(
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    trust_center_access_id TEXT NOT NULL REFERENCES trust_center_accesses(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    document_id TEXT REFERENCES documents(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    report_id TEXT REFERENCES reports(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    active BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (trust_center_access_id, document_id),
    UNIQUE (trust_center_access_id, report_id),
    CHECK ((document_id IS NOT NULL) != (report_id IS NOT NULL))
);
