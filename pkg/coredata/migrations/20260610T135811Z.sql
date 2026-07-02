-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
--
-- Permission to use, copy, modify, and/or distribute this software for any
-- purpose with or without fee is hereby granted, provided that the above
-- copyright notice and this permission notice appear in all copies.
--
-- THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
-- REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
-- AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
-- INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
-- LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
-- OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
-- PERFORMANCE OF THIS SOFTWARE.

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
