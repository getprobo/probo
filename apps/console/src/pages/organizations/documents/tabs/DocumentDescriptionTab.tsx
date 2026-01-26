import { Markdown } from "@probo/ui";
import { useOutletContext } from "react-router";

import type { DocumentLayoutQuery$data } from "#/__generated__/core/DocumentLayoutQuery.graphql";
import type { NodeOf } from "#/types";

export default function DocumentDescriptionTab() {
  const { version } = useOutletContext<{
    version: NodeOf<Extract<DocumentLayoutQuery$data["document"], { __typename: "Document" }>["versions"]>;
  }>();
  return (
    <div>
      <Markdown content={version.content} />
    </div>
  );
}
