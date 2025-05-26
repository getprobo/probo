import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";

type Props = {
    content: string;
};

export function Markdown({ content }: Props) {
    return (
        <div className="prose prose-neutral">
            <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                rehypePlugins={[rehypeRaw]}
            >
                {content}
            </ReactMarkdown>
        </div>
    );
}
