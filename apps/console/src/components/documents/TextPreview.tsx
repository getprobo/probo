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

import { Button, IconWarning, Markdown, Spinner } from "@probo/ui";
import { useEffect, useState } from "react";

import { consoleMarkdownImageOrigins } from "#/lib/markdownImageOrigins";

import { documentPreviewURL } from "./documentPreviewURL";

const MAX_PREVIEW_BYTES = 512 * 1024;

interface TextPreviewProps {
  src: string;
  emptyMessage: string;
  errorMessage: string;
  format: "markdown" | "text";
  retryLabel: string;
  truncatedMessage: string;
}

type PreviewState
  = | { status: "loading" }
    | { status: "error" }
    | {
      status: "ready";
      content: string;
      truncated: boolean;
    };

export function TextPreview({
  src,
  emptyMessage,
  errorMessage,
  format,
  retryLabel,
  truncatedMessage,
}: TextPreviewProps) {
  const [reloadKey, setReloadKey] = useState(0);
  const [preview, setPreview] = useState<PreviewState>({ status: "loading" });

  useEffect(() => {
    const abortController = new AbortController();

    async function loadPreview() {
      try {
        const response = await fetch(documentPreviewURL(src), {
          signal: abortController.signal,
        });
        if (!response.ok || response.body == null) {
          throw new Error(`Text download failed: ${response.status}`);
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let content = "";
        let loadedBytes = 0;
        let truncated = false;

        while (true) {
          const result = await reader.read();
          if (result.done) {
            break;
          }

          const remainingBytes = MAX_PREVIEW_BYTES - loadedBytes;
          if (result.value.byteLength > remainingBytes) {
            content += decoder.decode(
              result.value.subarray(0, remainingBytes),
              { stream: true },
            );
            truncated = true;
            await reader.cancel();
            break;
          }

          content += decoder.decode(result.value, { stream: true });
          loadedBytes += result.value.byteLength;
        }

        content += decoder.decode();
        if (truncated && content.endsWith("\uFFFD")) {
          content = content.slice(0, -1);
        }

        if (!abortController.signal.aborted) {
          setPreview({ status: "ready", content, truncated });
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
        <p className="text-center text-txt-secondary">{errorMessage}</p>
        <Button variant="secondary" onClick={retry}>
          {retryLabel}
        </Button>
      </div>
    );
  }

  if (preview.content.length === 0) {
    return (
      <p className="flex h-[50vh] items-center justify-center text-txt-secondary">
        {emptyMessage}
      </p>
    );
  }

  return (
    <div className="space-y-2">
      {format === "markdown"
        ? (
            <div className="max-h-[70vh] overflow-auto rounded-lg border border-border-low bg-tertiary p-6 [&_.prose]:max-w-none">
              <Markdown
                content={preview.content}
                allowedImageOrigins={consoleMarkdownImageOrigins()}
              />
            </div>
          )
        : (
            <pre className="max-h-[70vh] overflow-auto whitespace-pre-wrap break-words rounded-lg border border-border-low bg-tertiary p-4 font-mono text-sm text-txt-primary">
              {preview.content}
            </pre>
          )}
      {preview.truncated && (
        <p className="text-xs text-txt-tertiary">{truncatedMessage}</p>
      )}
    </div>
  );
}
