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
import { parse } from "papaparse";
import { useEffect, useState } from "react";

const MAX_DATA_ROWS = 100;

interface CSVPreviewProps {
  src: string;
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
  emptyMessage,
  errorMessage,
  retryLabel,
  truncatedMessage,
}: CSVPreviewProps) {
  const [reloadKey, setReloadKey] = useState(0);
  const [preview, setPreview] = useState<PreviewState>({ status: "loading" });

  useEffect(() => {
    const abortController = new AbortController();

    async function loadPreview() {
      try {
        const requestURL = new URL(src, window.location.href);
        if (
          import.meta.env.DEV
          && requestURL.pathname.startsWith("/api/files/v1/")
        ) {
          requestURL.protocol = window.location.protocol;
          requestURL.host = window.location.host;
        }
        const response = await fetch(requestURL, {
          signal: abortController.signal,
        });
        if (!response.ok) {
          throw new Error(`CSV download failed: ${response.status}`);
        }

        const text = await response.text();
        const result = parse<string[]>(text, {
          preview: MAX_DATA_ROWS + 1,
          skipEmptyLines: "greedy",
        });

        if (!abortController.signal.aborted) {
          setPreview({
            status: "ready",
            rows: result.data,
            truncated: result.meta.truncated,
          });
        }
      } catch (error) {
        if (
          !abortController.signal.aborted
          && !(error instanceof Error && error.name === "AbortError")
        ) {
          setPreview({ status: "error" });
        }
      }
    }

    void loadPreview();
    return () => {
      abortController.abort();
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

  const [header, ...rows] = preview.rows;
  return (
    <div className="space-y-2">
      <div className="max-h-[70vh] overflow-auto rounded-lg border border-border-low">
        <table className="w-full border-collapse text-left text-sm">
          <thead className="sticky top-0 z-10 bg-level-2 text-xs font-semibold text-txt-tertiary">
            <tr>
              {header.map((cell, index) => (
                <th
                  key={index}
                  scope="col"
                  className="max-w-80 whitespace-pre-wrap border-b border-r border-border-low px-3 py-2 last:border-r-0"
                >
                  {cell}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="bg-tertiary text-txt-primary">
            {rows.map((row, rowIndex) => (
              <tr
                key={rowIndex}
                className="border-b border-border-low last:border-b-0"
              >
                {header.map((_, columnIndex) => (
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
