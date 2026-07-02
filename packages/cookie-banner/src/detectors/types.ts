// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
