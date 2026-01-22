import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Card, PageHeader, Tbody, Th, Thead, Tr } from "@probo/ui";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { EmployeeDocumentsPageQuery } from "#/__generated__/core/EmployeeDocumentsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { DocumentRow } from "./_components/DocumentRow";

export const employeeDocumentsPageQuery = graphql`
  query EmployeeDocumentsPageQuery($organizationId: ID!) {
    viewer @required(action: THROW) {
      signableDocuments(
        organizationId: $organizationId
        first: 1000
        orderBy: { field: CREATED_AT, direction: DESC }
      ) @required(action: THROW) {
        edges @required(action: THROW) {
          node @required(action: THROW) {
            id
            ...DocumentRowFragment
          }
        }
      }
    }
  }
`;

export function EmployeeDocumentsPage(props: {
  queryRef: PreloadedQuery<EmployeeDocumentsPageQuery>;
}) {
  const { queryRef } = props;
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const {
    viewer: { signableDocuments },
  } = usePreloadedQuery<EmployeeDocumentsPageQuery>(
    employeeDocumentsPageQuery,
    queryRef,
  );

  const documents = signableDocuments.edges.map(edge => edge.node);

  usePageTitle(__("Documents"));

  return (
    <div className="space-y-6">
      <PageHeader title={__("Documents")} />
      {documents.length > 0
        ? (
            <Card>
              <table className="w-full">
                <Thead>
                  <Tr>
                    <Th className="min-w-0 pr-12">{__("Name")}</Th>
                    <Th className="w-48">{__("Type")}</Th>
                    <Th className="w-36">{__("Classification")}</Th>
                    <Th className="w-40">{__("Last update")}</Th>
                    <Th className="w-32">{__("Signed")}</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {documents.map(document => (
                    <DocumentRow
                      key={document.id}
                      fKey={document}
                      organizationId={organizationId}
                    />
                  ))}
                </Tbody>
              </table>
            </Card>
          )
        : (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-semibold mb-2">
                  {__("No documents yet")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {__("No documents have been requested for your signature.")}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}
