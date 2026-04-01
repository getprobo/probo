// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { CodeIcon, EyeIcon } from "@phosphor-icons/react";
import type { ReactNodeViewProps } from "@tiptap/react";
import { NodeViewContent, NodeViewWrapper } from "@tiptap/react";
import mermaid from "mermaid";
import { useEffect, useId, useState } from "react";

type MermaidMode = "code" | "preview";

function MermaidPreview({ chart }: { chart: string }) {
  const id = useId().replace(/:/g, "");
  const [svg, setSvg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const source = chart.trim();

  useEffect(() => {
    if (source.length === 0) return;

    let cancelled = false;

    mermaid.initialize({ startOnLoad: false, theme: "neutral" });

    mermaid
      .render(`mermaid-editor-${id}`, source)
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

  if (source.length === 0) {
    return (
      <div className="mermaid-empty">
        No diagram to display
      </div>
    );
  }

  if (error) {
    return (
      <div className="mermaid-error">
        {error}
      </div>
    );
  }

  if (!svg) {
    return (
      <div className="mermaid-empty">
        Rendering...
      </div>
    );
  }

  return (
    <div
      className="mermaid-preview"
      dangerouslySetInnerHTML={{ __html: svg }}
    />
  );
}

export function MermaidNodeView({ node }: ReactNodeViewProps) {
  const isMermaid = node.attrs.language === "mermaid";

  if (!isMermaid) {
    return (
      <NodeViewWrapper as="pre">
        <NodeViewContent<"code"> as="code" />
      </NodeViewWrapper>
    );
  }

  return <MermaidBlock node={node} />;
}

function MermaidBlock({ node }: { node: ReactNodeViewProps["node"] }) {
  const hasContent = node.textContent.trim().length > 0;
  const [mode, setMode] = useState<MermaidMode>(hasContent ? "preview" : "code");

  return (
    <NodeViewWrapper>
      <div className="mermaid-block">
        <div className="mermaid-toolbar">
          <button
            type="button"
            className={`mermaid-toolbar-btn ${mode === "code" ? "active" : ""}`}
            onClick={() => setMode("code")}
            onMouseDown={e => e.preventDefault()}
          >
            <CodeIcon size={14} weight="bold" />
            Code
          </button>
          <button
            type="button"
            className={`mermaid-toolbar-btn ${mode === "preview" ? "active" : ""}`}
            onClick={() => setMode("preview")}
            onMouseDown={e => e.preventDefault()}
          >
            <EyeIcon size={14} weight="bold" />
            Preview
          </button>
        </div>

        <pre className={mode === "preview" ? "hidden" : ""}>
          <NodeViewContent<"code"> as="code" />
        </pre>

        {mode === "preview" && (
          <MermaidPreview chart={node.textContent} />
        )}
      </div>
    </NodeViewWrapper>
  );
}
