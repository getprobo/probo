import ReactMarkdown from "react-markdown";
import rehypeRaw from "rehype-raw";
import remarkGfm from "remark-gfm";

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
          pre: ({ children, ...props }) => (
            <pre
              className="border border-border-solid rounded p-4 bg-transparent font-mono text-sm overflow-x-auto text-inherit"
              {...props}
            >
              {children}
            </pre>
          ),
          code: ({ children, ...props }) => (
            <code className="font-mono text-sm text-inherit" {...props}>
              {children}
            </code>
          ),
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
