// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { CodeIcon, EyeIcon } from "@phosphor-icons/react";
import type { ReactNodeViewProps } from "@tiptap/react";
import { NodeViewContent, NodeViewWrapper } from "@tiptap/react";
import { useEffect, useId, useState } from "react";

import { useToast } from "../Atoms/Toasts/Toasts";
import { mermaidRenderErrorToast, renderMermaidDiagram } from "../lib/mermaid";

type MermaidMode = "code" | "preview";

type MermaidRenderState = {
  source: string;
  svg: string | null;
  hasError: boolean;
};

function MermaidPreview({ chart }: { chart: string }) {
  const id = useId().replace(/:/g, "");
  const [renderState, setRenderState] = useState<MermaidRenderState>({
    source: "",
    svg: null,
    hasError: false,
  });
  const { toast } = useToast();

  const source = chart.trim();
  const svg = renderState.source === source ? renderState.svg : null;
  const hasError = renderState.source === source && renderState.hasError;

  useEffect(() => {
    if (source.length === 0) {
      return;
    }

    let cancelled = false;

    renderMermaidDiagram(`mermaid-editor-${id}`, source)
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

  if (source.length === 0) {
    return (
      <div className="mermaid-empty">
        No diagram to display
      </div>
    );
  }

  if (hasError) {
    return (
      <div className="mermaid-error">
        Unable to render diagram. Check the syntax and try again.
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
      <div className={`mermaid-block ${mode === "code" ? "editing" : ""}`}>
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
