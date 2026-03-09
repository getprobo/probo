import mermaid from "mermaid";
import { useEffect, useId, useState } from "react";

type Props = {
  chart: string;
};

export function MermaidDiagram({ chart }: Props) {
  const id = useId().replace(/:/g, "");
  const [svg, setSvg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    mermaid.initialize({ startOnLoad: false, theme: "neutral" });

    mermaid
      .render(`mermaid-${id}`, chart.trim())
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
  }, [chart, id]);

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
