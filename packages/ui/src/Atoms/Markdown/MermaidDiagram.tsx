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

import mermaid from "mermaid";
import { useEffect, useId, useState } from "react";

import { mermaidRenderConfig } from "../../mermaidConfig";

type Props = {
  chart: string;
};

export function MermaidDiagram({ chart }: Props) {
  const id = useId().replace(/:/g, "");
  const [svg, setSvg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const source = (chart ?? "").trim();

  useEffect(() => {
    if (!source) {
      return;
    }

    let cancelled = false;

    mermaid.initialize(mermaidRenderConfig);

    mermaid
      .render(`mermaid-${id}`, source)
      .then((result) => {
        if (!cancelled) {
          setSvg(result.svg);
          setError(null);
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : String(err));
        }
      });

    return () => {
      cancelled = true;
    };
  }, [source, id]);

  if (error) {
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
