// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { CodeIcon, EyeIcon } from "@phosphor-icons/react";
import type { ReactNodeViewProps } from "@tiptap/react";
import { NodeViewContent, NodeViewWrapper } from "@tiptap/react";
import { useState } from "react";

import { MermaidDiagram } from "../Atoms/Markdown/MermaidDiagram";

type MermaidMode = "code" | "preview";

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
          <MermaidDiagram chart={node.textContent.trim()} />
        )}
      </div>
    </NodeViewWrapper>
  );
}
