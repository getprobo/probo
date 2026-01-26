import { useTranslate } from "@probo/i18n";
import { Badge } from "@probo/ui";
import { clsx } from "clsx";
import { useFragment } from "react-relay";
import { Link, useLocation, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type {
  DocumentVersionsDropdownItemFragment$key,
} from "#/__generated__/core/DocumentVersionsDropdownItemFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
  fragment DocumentVersionsDropdownItemFragment on DocumentVersion {
    id
    version
    status
    publishedAt
    updatedAt
  }
`;

export function DocumentVersionsDropdownItem(props: {
  fKey: DocumentVersionsDropdownItemFragment$key;
  active?: boolean;
}) {
  const { fKey, active } = props;

  const { dateTimeFormat, __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { documentId } = useParams();
  if (!documentId) {
    throw new Error(":documentId route param missing");
  }

  const version = useFragment<DocumentVersionsDropdownItemFragment$key>(fragment, fKey);

  const suffix = useLocation().pathname.split("/").at(-1);

  return (
    <Link
      to={`/organizations/${organizationId}/documents/${documentId}/versions/${version.id}/${suffix}`}
      className="flex items-center gap-2 py-2 px-[10px] w-full hover:bg-tertiary-hover cursor-pointer rounded"
      {...props}
    >
      <div className="flex gap-3 w-full overflow-hidden">
        <div
          className={clsx(
            "shrink-0 flex items-center justify-center size-10",
            active && "bg-active rounded",
          )}
        >
          <div className="text-base text-txt-primary whitespace-nowrap font-bold text-center">
            {version.version}
          </div>
        </div>
        <div className="flex-1 space-y-[2px] overflow-hidden">
          <div className="flex items-center gap-2 overflow-hidden">
            {version.status === "DRAFT" && (
              <Badge variant="neutral" size="sm">
                {__("Draft")}
              </Badge>
            )}
          </div>
          <div className="text-xs text-txt-secondary whitespace-nowrap overflow-hidden text-ellipsis">
            {dateTimeFormat(version.publishedAt ?? version.updatedAt)}
          </div>
        </div>
      </div>
    </Link>
  );
}
