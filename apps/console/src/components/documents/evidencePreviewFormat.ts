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

export type EvidencePreviewFormat
  = | "csv"
    | "image"
    | "json"
    | "markdown"
    | "pdf"
    | "text"
    | "unavailable"
    | "uri";

export function evidencePreviewFormat(
  rawMimeType: string | null | undefined,
  rawFileName: string | null | undefined,
): EvidencePreviewFormat {
  const mimeType = rawMimeType?.split(";")[0]?.trim().toLowerCase() ?? "";
  const fileName = rawFileName?.toLowerCase() ?? "";

  if (mimeType === "text/uri-list" || mimeType === "text/uri") {
    return "uri";
  }
  if (mimeType.startsWith("image/")) {
    return "image";
  }
  if (mimeType.includes("pdf")) {
    return "pdf";
  }
  if (
    mimeType === "text/csv"
    || mimeType === "application/csv"
    || fileName.endsWith(".csv")
  ) {
    return "csv";
  }
  if (
    mimeType === "application/json"
    || mimeType === "text/json"
    || fileName.endsWith(".json")
  ) {
    return "json";
  }
  if (
    mimeType === "text/markdown"
    || fileName.endsWith(".md")
    || fileName.endsWith(".markdown")
  ) {
    return "markdown";
  }
  if (
    mimeType.startsWith("text/")
    || fileName.endsWith(".txt")
    || fileName.endsWith(".log")
  ) {
    return "text";
  }

  return "unavailable";
}
