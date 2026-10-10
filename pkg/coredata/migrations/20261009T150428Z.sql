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

CREATE TABLE attachments (
    organization_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,

    file_id TEXT NOT NULL REFERENCES files(id) ON UPDATE CASCADE ON DELETE RESTRICT,

    document_version_id TEXT REFERENCES document_versions(id) ON UPDATE CASCADE ON DELETE CASCADE,
    task_id TEXT REFERENCES tasks(id) ON UPDATE CASCADE ON DELETE CASCADE,
    task_comment_id TEXT REFERENCES task_comments(id) ON UPDATE CASCADE ON DELETE CASCADE,
    risk_analysis_id TEXT REFERENCES risk_analyses(id) ON UPDATE CASCADE ON DELETE CASCADE,
    parent_id TEXT GENERATED ALWAYS AS (
        COALESCE(document_version_id, task_id, task_comment_id, risk_analysis_id)
    ) STORED NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT attachments_one_parent CHECK (
        (
            (document_version_id IS NOT NULL)::integer
            + (task_id IS NOT NULL)::integer
            + (task_comment_id IS NOT NULL)::integer
            + (risk_analysis_id IS NOT NULL)::integer
        ) = 1
    )
);

CREATE UNIQUE INDEX attachments_parent_key
    ON attachments (file_id, parent_id);
