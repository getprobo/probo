// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

export const acceptDocument = {
  "application/pdf": [".pdf"],
  "application/msword": [".doc"],
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document": [".docx"],
  "application/vnd.oasis.opendocument.text": [".odt"],
} satisfies Record<string, string[]>;

export const acceptSpreadsheet = {
  "application/vnd.ms-excel": [".xls"],
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [".xlsx"],
  "application/vnd.oasis.opendocument.spreadsheet": [".ods"],
} satisfies Record<string, string[]>;

export const acceptPresentation = {
  "application/vnd.ms-powerpoint": [".ppt"],
  "application/vnd.openxmlformats-officedocument.presentationml.presentation": [".pptx"],
  "application/vnd.oasis.opendocument.presentation": [".odp"],
} satisfies Record<string, string[]>;

export const acceptText = {
  "text/markdown": [".md"],
  "text/plain": [".txt"],
  "text/x-log": [".log"],
  "text/uri-list": [".uri"],
  "text/uri-list; charset=utf-8": [".uri"],
} satisfies Record<string, string[]>;

export const acceptImage = {
  "image/jpeg": [".jpg", ".jpeg"],
  "image/png": [".png"],
  "image/svg+xml": [".svg"],
  "image/webp": [".webp"],
} satisfies Record<string, string[]>;

export const acceptData = {
  "application/yaml": [".yaml", ".yml"],
  "application/json": [".json"],
  "text/yaml": [".yaml", ".yml"],
  "text/json": [".json"],
  "text/csv": [".csv"],
  "application/csv": [".csv"],
} satisfies Record<string, string[]>;

export const acceptVideo = {
  "video/mp4": [".mp4"],
  "video/mpeg": [".mpeg", ".mpg"],
  "video/quicktime": [".mov"],
  "video/x-msvideo": [".avi"],
  "video/webm": [".webm"],
} satisfies Record<string, string[]>;

export const acceptAll = {
  ...acceptDocument,
  ...acceptSpreadsheet,
  ...acceptPresentation,
  ...acceptText,
  ...acceptImage,
  ...acceptData,
  ...acceptVideo,
} satisfies Record<string, string[]>;
