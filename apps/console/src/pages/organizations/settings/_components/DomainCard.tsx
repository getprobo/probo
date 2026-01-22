import {
  getCustomDomainStatusBadgeLabel,
  getCustomDomainStatusBadgeVariant,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Button, Card } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DomainCardFragment$key } from "/__generated__/core/DomainCardFragment.graphql";

import { DeleteDomainDialog } from "./DeleteDomainDialog";
import { DomainDialog } from "./DomainDialog";

const fragment = graphql`
  fragment DomainCardFragment on CustomDomain {
    domain
    sslStatus
    canDelete: permission(action: "core:custom-domain:delete")
    ...DomainDialogFragment
  }
`;

export function DomainCard(props: { fKey: DomainCardFragment$key }) {
  const { fKey } = props;

  const { __ } = useTranslate();

  const domain = useFragment<DomainCardFragment$key>(fragment, fKey);

  return (
    <Card>
      <div className="p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div>
              <div className="font-medium mb-1">{domain.domain}</div>
              <div className="text-sm text-txt-secondary">
                {domain.sslStatus === "ACTIVE"
                  ? __("Verified")
                  : __("Pending verification")}
              </div>
            </div>
            <Badge
              variant={getCustomDomainStatusBadgeVariant(domain.sslStatus)}
            >
              {getCustomDomainStatusBadgeLabel(domain.sslStatus, __)}
            </Badge>
          </div>

          <div className="flex items-center gap-2">
            <DomainDialog fKey={domain}>
              <Button variant="secondary">{__("View Details")}</Button>
            </DomainDialog>

            {domain.canDelete && (
              <DeleteDomainDialog domain={domain.domain}>
                <Button variant="danger">{__("Delete")}</Button>
              </DeleteDomainDialog>
            )}
          </div>
        </div>
      </div>
    </Card>
  );
}
