import { useTranslate } from "@probo/i18n";
import { Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageDocumentListFragment$key } from "#/__generated__/core/CompliancePageDocumentListFragment.graphql";

import { CompliancePageDocumentListItem } from "./CompliancePageDocumentListItem";

const fragment = graphql`
  fragment CompliancePageDocumentListFragment on Organization {
    compliancePage: trustCenter @required(action: THROW) {
      ...CompliancePageDocumentListItem_compliancePageFragment
    }
    documents(first: 100) {
      edges {
        node {
          id
          currentPublishedVersion
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
  const publishedDocuments = documents.edges.filter(({ node }) => node.currentPublishedVersion != null);

  return (
    <div className="space-y-[10px]">
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Type")}</Th>
            <Th>{__("Visibility")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {publishedDocuments.length === 0 && (
            <Tr>
              <Td colSpan={3} className="text-center text-txt-secondary">
                {__("No documents available")}
              </Td>
            </Tr>
          )}
          {publishedDocuments.map(({ node: document }) => (
            <CompliancePageDocumentListItem
              key={document.id}
              compliancePageFragmentRef={compliancePage}
              documentFragmentRef={document}
            />
          ))}
        </Tbody>
      </Table>
    </div>
  );
};
