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

ALTER TYPE webhook_event_type ADD VALUE 'document:created';
ALTER TYPE webhook_event_type ADD VALUE 'document:updated';
ALTER TYPE webhook_event_type ADD VALUE 'document:archived';
ALTER TYPE webhook_event_type ADD VALUE 'document:unarchived';
ALTER TYPE webhook_event_type ADD VALUE 'document:deleted';

ALTER TYPE webhook_event_type ADD VALUE 'document-version:created';
ALTER TYPE webhook_event_type ADD VALUE 'document-version:updated';
ALTER TYPE webhook_event_type ADD VALUE 'document-version:published';
ALTER TYPE webhook_event_type ADD VALUE 'document-version:rejected';
ALTER TYPE webhook_event_type ADD VALUE 'document-version:deleted';

ALTER TYPE webhook_event_type ADD VALUE 'document-version-signature:requested';
ALTER TYPE webhook_event_type ADD VALUE 'document-version-signature:signed';
ALTER TYPE webhook_event_type ADD VALUE 'document-version-signature:cancelled';

ALTER TYPE webhook_event_type ADD VALUE 'document-version-approval-quorum:requested';
ALTER TYPE webhook_event_type ADD VALUE 'document-version-approval-quorum:updated';
ALTER TYPE webhook_event_type ADD VALUE 'document-version-approval-quorum:approved';
ALTER TYPE webhook_event_type ADD VALUE 'document-version-approval-quorum:rejected';
ALTER TYPE webhook_event_type ADD VALUE 'document-version-approval-quorum:voided';
