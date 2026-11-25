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

import type { DocumentMode } from "./useDocumentModes";

// The document page shows a previous version while this param holds its id,
// and the live document otherwise.
export const documentVersionParam = "version";
export const documentModeParam = "mode";
export const documentColumnParam = "view";

// The column beside the document: the open comments, the comment history, or
// the versions.
export type DocumentColumn = "open" | "all" | "versions";

function isDocumentColumn(value: string): value is DocumentColumn {
  return value === "open" || value === "all" || value === "versions";
}

export function readDocumentVersion(params: URLSearchParams): string | null {
  return params.get(documentVersionParam);
}

// Past versions are read only, so a version takes precedence over the mode.
export function readDocumentMode(params: URLSearchParams): DocumentMode {
  if (params.has(documentVersionParam)) {
    return "viewing";
  }
  const mode = params.get(documentModeParam);
  return mode === "editing" || mode === "suggesting" ? mode : "viewing";
}

// A link to a version opens the versions beside it unless it names a column.
function defaultDocumentColumn(params: URLSearchParams): DocumentColumn {
  return params.has(documentVersionParam) ? "versions" : "open";
}

export function readDocumentColumn(params: URLSearchParams): DocumentColumn {
  const column = params.get(documentColumnParam);
  return column != null && isDocumentColumn(column) ? column : defaultDocumentColumn(params);
}

function writeDocumentMode(params: URLSearchParams, mode: DocumentMode) {
  if (mode === "viewing") {
    params.delete(documentModeParam);
  } else {
    params.set(documentModeParam, mode);
  }
}

function writeDocumentColumn(params: URLSearchParams, column: DocumentColumn) {
  if (column === defaultDocumentColumn(params)) {
    params.delete(documentColumnParam);
  } else {
    params.set(documentColumnParam, column);
  }
}

// Edits apply to the live document, so editing or suggesting leaves the
// version and its list.
export function withDocumentMode(params: URLSearchParams, mode: DocumentMode): URLSearchParams {
  const next = new URLSearchParams(params);
  const column = readDocumentColumn(params);
  writeDocumentMode(next, mode);
  if (mode !== "viewing") {
    next.delete(documentVersionParam);
    writeDocumentColumn(next, column === "versions" ? "open" : column);
  }
  return next;
}

export function withDocumentVersion(params: URLSearchParams, versionId: string | null): URLSearchParams {
  const next = new URLSearchParams(params);
  const column = readDocumentColumn(params);
  if (versionId == null) {
    next.delete(documentVersionParam);
  } else {
    next.set(documentVersionParam, versionId);
    next.delete(documentModeParam);
  }
  writeDocumentColumn(next, column);
  return next;
}

export function withDocumentColumn(params: URLSearchParams, column: DocumentColumn): URLSearchParams {
  const next = new URLSearchParams(params);
  writeDocumentColumn(next, column);
  return next;
}

// Drops what the page does not show: a mode the viewer may not use, a mode
// beside a version, and values that restate a default.
export function normalizeDocumentSearch(
  params: URLSearchParams,
  modes: readonly DocumentMode[],
): URLSearchParams {
  const next = new URLSearchParams(params);
  const mode = readDocumentMode(params);
  writeDocumentMode(next, modes.includes(mode) ? mode : "viewing");
  writeDocumentColumn(next, readDocumentColumn(params));
  return next;
}

// The document list keeps its filters in the search the details page was
// opened with. The version, mode, and column belong to the details page only.
export function documentListSearch(search: string): string {
  const params = new URLSearchParams(search);
  params.delete(documentVersionParam);
  params.delete(documentModeParam);
  params.delete(documentColumnParam);
  const next = params.toString();
  return next === "" ? "" : `?${next}`;
}
