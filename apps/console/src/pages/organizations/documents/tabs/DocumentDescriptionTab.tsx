import { useOutletContext } from "react-router";
import type { DocumentDetailPageDocumentFragment$data } from "/__generated__/core/DocumentDetailPageDocumentFragment.graphql";
import type { NodeOf } from "/types";
import { Markdown } from "@probo/ui";

export default function DocumentDescriptionTab() {
  const { version } = useOutletContext<{
    version: NodeOf<DocumentDetailPageDocumentFragment$data["versions"]>;
  }>();
  return (
    <div>
      <Markdown content={version.content} />
    </div>
  );
}
