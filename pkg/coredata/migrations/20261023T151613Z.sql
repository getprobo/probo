-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy
-- of this software and associated documentation files (the "Software"), to deal
-- in the Software without restriction, including without limitation the rights
-- to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
-- copies of the Software, and to permit persons to whom the Software is
-- furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in
-- all copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
-- FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
-- AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
-- LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
-- OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
-- SOFTWARE.

CREATE TABLE document_suggestions (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    document_id TEXT NOT NULL REFERENCES documents(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    author_profile_id TEXT NOT NULL REFERENCES iam_membership_profiles(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    status TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by_profile_id TEXT REFERENCES iam_membership_profiles(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE document_comment_threads (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    document_id TEXT NOT NULL REFERENCES documents(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    suggestion_id TEXT REFERENCES document_suggestions(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    anchor_start BYTEA,
    anchor_end BYTEA,
    anchor_quote TEXT NOT NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by_profile_id TEXT REFERENCES iam_membership_profiles(id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX document_comment_threads_open_suggestion_key
    ON document_comment_threads (suggestion_id)
    WHERE resolved_at IS NULL AND suggestion_id IS NOT NULL;

CREATE TABLE document_comments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    thread_id TEXT NOT NULL REFERENCES document_comment_threads(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    author_profile_id TEXT NOT NULL REFERENCES iam_membership_profiles(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
