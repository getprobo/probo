import { useOutletContext } from "react-router";
import type { DocumentDetailPageDocumentFragment$data } from "../__generated__/DocumentDetailPageDocumentFragment.graphql";
import type { NodeOf } from "/types";
import { Markdown } from "@probo/ui";

export default function DocumentDescriptionTab() {
  const { lastVersion } = useOutletContext<{
    lastVersion: NodeOf<DocumentDetailPageDocumentFragment$data["versions"]>;
  }>();
  return (
    <div>
      <Markdown content={lastVersion.content} />
    </div>
  );
}
