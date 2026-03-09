import { isValidElement } from "react";
import ReactMarkdown from "react-markdown";
import rehypeRaw from "rehype-raw";
import remarkGfm from "remark-gfm";

import { MermaidDiagram } from "./MermaidDiagram";

type Props = {
  content: string;
};

export function Markdown({ content }: Props) {
  return (
    <div className="prose prose-neutral">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeRaw]}
        components={{
          a: ({ href, children, ...props }) => (
            <a
              href={href}
              target="_blank"
              rel="noopener noreferrer"
              {...props}
            >
              {children}
            </a>
          ),
          pre: ({ children, ...props }) => {
            const child = isValidElement<{
              className?: string;
              children?: string;
            }>(children)
              ? children
              : null;

            if (
              child?.type === "code"
              && child.props.className === "language-mermaid"
              && typeof child.props.children === "string"
            ) {
              return <MermaidDiagram chart={child.props.children} />;
            }

            return (
              <pre
                className="border border-border-solid rounded p-4 bg-transparent font-mono text-sm overflow-x-auto text-inherit"
                {...props}
              >
                {children}
              </pre>
            );
          },
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
