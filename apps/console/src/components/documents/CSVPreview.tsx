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

import { Button, IconWarning, Spinner } from "@probo/ui";
import { parse, type Parser } from "papaparse";
import { useEffect, useState } from "react";

const MAX_PREVIEW_ROWS = 100;
const PREVIEW_CHUNK_SIZE = 64 * 1024;

interface CSVPreviewProps {
  src: string;
  columnLabel: string;
  emptyMessage: string;
  errorMessage: string;
  retryLabel: string;
  truncatedMessage: string;
}

type PreviewState
  = | { status: "loading" }
    | { status: "error" }
    | {
      status: "ready";
      rows: string[][];
      truncated: boolean;
    };

export function CSVPreview({
  src,
  columnLabel,
  emptyMessage,
  errorMessage,
  retryLabel,
  truncatedMessage,
}: CSVPreviewProps) {
  const [reloadKey, setReloadKey] = useState(0);
  const [preview, setPreview] = useState<PreviewState>({ status: "loading" });

  useEffect(() => {
    let activeParser: Parser | undefined;
    let cancelled = false;
    let failed = false;

    const requestURL = new URL(src, window.location.href);
    if (
      import.meta.env.DEV
        && requestURL.pathname.startsWith("/api/files/v1/")
    ) {
      requestURL.protocol = window.location.protocol;
      requestURL.host = window.location.host;
    }

    const rows: string[][] = [];
    parse<string[]>(requestURL.href, {
      chunkSize: PREVIEW_CHUNK_SIZE,
      delimiter: ",",
      download: true,
      preview: MAX_PREVIEW_ROWS + 1,
      skipEmptyLines: "greedy",
      chunk(result, parser) {
        activeParser = parser;
        if (cancelled) {
          parser.abort();
          return;
        }

        if (result.errors.length > 0) {
          failed = true;
          parser.abort();
          setPreview({ status: "error" });
          return;
        }

        rows.push(...result.data);
      },
      complete() {
        activeParser = undefined;
        if (!cancelled && !failed) {
          setPreview({
            status: "ready",
            rows: rows.slice(0, MAX_PREVIEW_ROWS),
            truncated: rows.length > MAX_PREVIEW_ROWS,
          });
        }
      },
      error() {
        activeParser = undefined;
        if (!cancelled) {
          setPreview({ status: "error" });
        }
      },
    });

    return () => {
      cancelled = true;
      activeParser?.abort();
    };
  }, [reloadKey, src]);

  function retry() {
    setPreview({ status: "loading" });
    setReloadKey(key => key + 1);
  }

  if (preview.status === "loading") {
    return (
      <div className="flex h-[50vh] items-center justify-center">
        <Spinner />
      </div>
    );
  }

  if (preview.status === "error") {
    return (
      <div className="flex h-[50vh] flex-col items-center justify-center gap-3">
        <IconWarning size={20} />
        <p className="text-txt-secondary text-center">{errorMessage}</p>
        <Button
          variant="secondary"
          onClick={retry}
        >
          {retryLabel}
        </Button>
      </div>
    );
  }

  if (preview.rows.length === 0) {
    return (
      <p className="flex h-[50vh] items-center justify-center text-txt-secondary">
        {emptyMessage}
      </p>
    );
  }

  const columnCount = preview.rows.reduce(
    (count, row) => Math.max(count, row.length),
    0,
  );
  return (
    <div className="space-y-2">
      <div className="max-h-[70vh] overflow-auto rounded-lg border border-border-low">
        <table className="w-full border-collapse text-left text-sm">
          <thead className="sticky top-0 z-10 bg-level-2 text-xs font-semibold text-txt-tertiary">
            <tr>
              {Array.from({ length: columnCount }, (_, index) => (
                <th
                  key={index}
                  scope="col"
                  className="max-w-80 whitespace-pre-wrap border-b border-r border-border-low px-3 py-2 last:border-r-0"
                >
                  {columnLabel}
                  {" "}
                  {index + 1}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="bg-tertiary text-txt-primary">
            {preview.rows.map((row, rowIndex) => (
              <tr
                key={rowIndex}
                className="border-b border-border-low last:border-b-0"
              >
                {Array.from({ length: columnCount }, (_, columnIndex) => (
                  <td
                    key={columnIndex}
                    className="max-w-80 whitespace-pre-wrap break-words border-r border-border-low px-3 py-2 align-top last:border-r-0"
                  >
                    {row[columnIndex] ?? ""}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {preview.truncated && (
        <p className="text-xs text-txt-tertiary">{truncatedMessage}</p>
      )}
    </div>
  );
}
