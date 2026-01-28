import { useTranslate } from "@probo/i18n";
import { Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageAuditListFragment$key } from "#/__generated__/core/CompliancePageAuditListFragment.graphql";

import { CompliancePageAuditListItem } from "./CompliancePageAuditListItem";

const fragment = graphql`
  fragment CompliancePageAuditListFragment on Organization {
    compliancePage: trustCenter @required(action: THROW) {
      ...CompliancePageAuditListItem_compliancePageFragment
    }
    audits(first: 100) {
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

  const { audits, compliancePage } = useFragment<CompliancePageAuditListFragment$key>(fragment, fragmentRef);

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
    </div>
  );
}
