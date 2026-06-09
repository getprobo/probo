// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

// The string-literal unions below mirror the server's enum values
// accepted by `POST /{bannerID}/report` in
// pkg/server/api/cookiebanner/v1/handler.go. Any change here must be
// matched server-side or the request will be rejected.

export type CookieSource = "script" | "pre-existing" | "http" | "extension";

export type StorageSource = "script" | "pre-existing" | "extension";

export type StorageType =
  | "local_storage"
  | "session_storage"
  | "indexed_db"
  | "cache_storage";

export type ResourceType =
  | "script"
  | "iframe"
  | "image"
  | "stylesheet"
  | "font"
  | "beacon"
  | "fetch"
  | "media"
  | "service_worker";

export interface DetectedCookieEntry {
  name: string;
  max_age_seconds: number | null;
  source: CookieSource;
  initiator_url?: string;
}

export interface DetectedStorageEntry {
  key: string;
  storage_type: StorageType;
  value_size: number | null;
  source: StorageSource;
  initiator_url?: string;
}

export interface DetectedResourceEntry {
  url: string;
  resource_type: ResourceType;
}
