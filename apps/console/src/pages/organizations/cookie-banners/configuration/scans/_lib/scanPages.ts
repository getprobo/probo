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

export interface ScanPageRow {
  url: string;
  scanned: boolean;
}

export function scanPageRows(
  visited: readonly string[],
  queue: readonly { readonly url: string }[],
): ScanPageRow[] {
  const seen = new Set<string>();
  const rows: ScanPageRow[] = [];

  for (const url of visited) {
    if (url === "" || seen.has(url)) {
      continue;
    }
    seen.add(url);
    rows.push({ url, scanned: true });
  }

  for (const item of queue) {
    if (item.url === "" || seen.has(item.url)) {
      continue;
    }
    seen.add(item.url);
    rows.push({ url: item.url, scanned: false });
  }

  return rows;
}

export function scanPageTotals(
  rows: readonly ScanPageRow[],
  pagesScanned: number,
): { scanned: number; total: number } {
  let queued = 0;
  for (const row of rows) {
    if (!row.scanned) {
      queued += 1;
    }
  }

  return { scanned: pagesScanned, total: pagesScanned + queued };
}
