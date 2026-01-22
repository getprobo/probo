import { useOutletContext } from "react-router";
import { Markdown } from "@probo/ui";

import type { DocumentDetailPageDocumentFragment$data } from "/__generated__/core/DocumentDetailPageDocumentFragment.graphql";
import type { NodeOf } from "/types";

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
