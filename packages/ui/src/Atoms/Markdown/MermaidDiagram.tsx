// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { useEffect, useId, useState } from "react";

import { mermaidRenderErrorToast, renderMermaidDiagram } from "../../lib/mermaid";
import { useToast } from "../Toasts/Toasts";

type Props = {
  chart: string;
};

type MermaidRenderState = {
  source: string;
  svg: string | null;
  hasError: boolean;
};

export function MermaidDiagram({ chart }: Props) {
  const id = useId().replace(/:/g, "");
  const [renderState, setRenderState] = useState<MermaidRenderState>({
    source: "",
    svg: null,
    hasError: false,
  });
  const { toast } = useToast();

  const source = (chart ?? "").trim();
  const svg = renderState.source === source ? renderState.svg : null;
  const hasError = renderState.source === source && renderState.hasError;

  useEffect(() => {
    if (!source) {
      return;
    }

    let cancelled = false;

    renderMermaidDiagram(`mermaid-${id}`, source)
      .then((result) => {
        if (!cancelled) {
          setRenderState({
            source,
            svg: result.svg,
            hasError: false,
          });
        }
      })
      .catch(() => {
        if (!cancelled) {
          setRenderState({
            source,
            svg: null,
            hasError: true,
          });
          toast(mermaidRenderErrorToast);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [source, id, toast]);

  if (hasError) {
    return (
      <pre className="border border-border-solid rounded p-4 bg-transparent font-mono text-sm overflow-x-auto text-inherit">
        <code>{chart}</code>
      </pre>
    );
  }

  return (
    <div
      className="flex justify-center my-4"
      dangerouslySetInnerHTML={svg ? { __html: svg } : undefined}
    />
  );
}
