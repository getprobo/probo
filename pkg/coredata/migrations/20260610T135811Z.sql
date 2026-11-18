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

-- Clear stale tracker initiator attribution that wrongly pointed at the
-- @probo/cookie-banner SDK bundle. A pre-fix SDK bug walked its own bundle
-- frame when computing the initiator, so cookies/storage written by third
-- parties (or by malware / browser extensions) were attributed to
-- cookie-banner.iife.js. The SDK now excludes its own bundle, but the report
-- upsert keeps the existing value on re-detection
-- (initiator_url = COALESCE(new, old)), so rows whose corrected initiator is
-- NULL would otherwise keep the stale value forever. NULL them here; genuine
-- third-party rows repopulate with the correct initiator on next detection.
UPDATE detected_trackers
SET initiator_url    = NULL,
    initiator_domain = NULL,
    updated_at       = now()
WHERE initiator_url LIKE '%@probo/cookie-banner%'
   OR initiator_url LIKE '%cookie-banner.iife.js'
   OR initiator_url LIKE '%cookie-banner.mjs';
