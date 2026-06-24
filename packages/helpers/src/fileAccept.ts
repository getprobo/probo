// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
