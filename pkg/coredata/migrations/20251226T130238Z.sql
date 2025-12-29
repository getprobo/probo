CREATE TYPE rights_request_type AS ENUM (
    'ACCESS',
    'DELETION',
    'PORTABILITY'
);

CREATE TYPE rights_request_state AS ENUM (
    'TODO',
    'IN_PROGRESS',
    'DONE'
);

CREATE TABLE rights_requests (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL,

    request_type rights_request_type NOT NULL,
    request_state rights_request_state NOT NULL,

    data_subject TEXT,
    contact TEXT,
    details TEXT,

    deadline DATE,

    action_taken TEXT,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT rights_requests_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);
