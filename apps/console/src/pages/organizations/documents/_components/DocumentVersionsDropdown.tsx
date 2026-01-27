import { useTranslate } from "@probo/i18n";
import { Button, Dropdown, IconChevronDown, IconClock } from "@probo/ui";
import { Suspense } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { DocumentVersionsDropdownMenuQuery } from "#/__generated__/core/DocumentVersionsDropdownMenuQuery.graphql";

import { DocumentVersionsDropdownMenu, documentVersionsDropdownMenuQuery } from "./DocumentVersionsDropdownMenu";

export function DocumentVersionsDropdown() {
  const { documentId, versionId } = useParams();
  if (!documentId) {
    throw new Error(":documentId missing in route params");
  }
  const [queryRef, loadQuery] = useQueryLoader<DocumentVersionsDropdownMenuQuery>(documentVersionsDropdownMenuQuery);

  const { __ } = useTranslate();

  return (
    <Dropdown
      onOpenChange={open => open && !queryRef && loadQuery({ documentId, versionId: versionId ?? "", versionSpecified: !!versionId })}
      toggle={(
        <Button icon={IconClock} variant="secondary">
          {__("Version history")}
          <IconChevronDown size={12} />
        </Button>
      )}
    >
      <Suspense>
        {queryRef
          && (
            <DocumentVersionsDropdownMenu queryRef={queryRef} />
          )}
      </Suspense>
    </Dropdown>
  );
}
