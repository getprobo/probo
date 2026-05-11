-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

-- Extend tracker_resource_type to cover everything the in-page
-- PerformanceObserver can attribute: tracking pixels, fetch/XHR
-- call-homes, beacons, web fonts, cross-origin stylesheets, and
-- video/audio embeds.
ALTER TYPE tracker_resource_type ADD VALUE IF NOT EXISTS 'IMAGE';
ALTER TYPE tracker_resource_type ADD VALUE IF NOT EXISTS 'STYLESHEET';
ALTER TYPE tracker_resource_type ADD VALUE IF NOT EXISTS 'FONT';
ALTER TYPE tracker_resource_type ADD VALUE IF NOT EXISTS 'BEACON';
ALTER TYPE tracker_resource_type ADD VALUE IF NOT EXISTS 'FETCH';
ALTER TYPE tracker_resource_type ADD VALUE IF NOT EXISTS 'MEDIA';
