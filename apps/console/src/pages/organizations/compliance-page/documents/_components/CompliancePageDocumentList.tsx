import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, IconChevronDown, Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageDocumentListFragment$key } from "#/__generated__/core/CompliancePageDocumentListFragment.graphql";

import { CompliancePageDocumentListItem } from "./CompliancePageDocumentListItem";

const fragment = graphql`
  fragment CompliancePageDocumentListFragment on Organization {
    compliancePage: trustCenter @required(action: THROW) {
      ...CompliancePageDocumentListItem_compliancePageFragment
    }
    documents(first: 1000) {
      edges {
        node {
          id
          ...CompliancePageDocumentListItem_documentFragment
        }
      }
    }
  }
`;

export function CompliancePageDocumentList(props: { fragmentRef: CompliancePageDocumentListFragment$key }) {
  const { fragmentRef } = props;

  const { __ } = useTranslate();

  const { compliancePage, documents } = useFragment<CompliancePageDocumentListFragment$key>(fragment, fragmentRef);

  const [limit, setLimit] = useState<number | null>(100);
  const showMoreButton = limit !== null && documents.edges.length > limit;

  return (
    <div className="space-y-[10px]">
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Type")}</Th>
            <Th>{__("State")}</Th>
            <Th>{__("Visibility")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {documents.edges.length === 0 && (
            <Tr>
              <Td colSpan={5} className="text-center text-txt-secondary">
                {__("No documents available")}
              </Td>
            </Tr>
          )}
          {documents.edges.map(({ node: document }) => (
            <CompliancePageDocumentListItem
              key={document.id}
              compliancePageFragmentRef={compliancePage}
              documentFragmentRef={document}
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
          {sprintf(__("Show %s more"), documents.edges.length - limit)}
        </Button>
      )}
    </div>
  );
};
