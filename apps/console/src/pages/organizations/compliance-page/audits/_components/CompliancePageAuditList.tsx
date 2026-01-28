import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, IconChevronDown, Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageAuditListFragment$key } from "#/__generated__/core/CompliancePageAuditListFragment.graphql";

import { CompliancePageAuditListItem } from "./CompliancePageAuditListItem";

const fragment = graphql`
  fragment CompliancePageAuditListFragment on Organization {
    compliancePage: trustCenter @required(action: THROW) {
      ...CompliancePageAuditListItem_compliancePageFragment
    }
    audits(first: 1000) {
      edges {
        node {
          id
          ...CompliancePageAuditListItem_auditFragment
        }
      }
    }
  }
`;

export function CompliancePageAuditList(props: { fragmentRef: CompliancePageAuditListFragment$key }) {
  const { fragmentRef } = props;

  const { __ } = useTranslate();
  const [limit, setLimit] = useState<number | null>(100);

  const { audits, compliancePage } = useFragment<CompliancePageAuditListFragment$key>(fragment, fragmentRef);

  const showMoreButton = limit !== null && audits.edges.length > limit;

  return (
    <div className="space-y-[10px]">
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Framework")}</Th>
            <Th>{__("Name")}</Th>
            <Th>{__("Valid Until")}</Th>
            <Th>{__("State")}</Th>
            <Th>{__("Visibility")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {audits.edges.length === 0 && (
            <Tr>
              <Td colSpan={6} className="text-center text-txt-secondary">
                {__("No audits available")}
              </Td>
            </Tr>
          )}
          {audits.edges.map(({ node: audit }) => (
            <CompliancePageAuditListItem
              key={audit.id}
              auditFragmentRef={audit}
              compliancePageFragmentRef={compliancePage}
            />
          ))}
        </Tbody>
      </Table>
      {showMoreButton && (
        <Button
          variant="tertiary"
          onClick={() => setLimit(null)}
          className="mt-3 mx-auto"
          icon={IconChevronDown}
        >
          {sprintf(__("Show %s more"), audits.edges.length - limit)}
        </Button>
      )}
    </div>
  );
}
