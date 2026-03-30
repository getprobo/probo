// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Card, Tbody, Th, Thead, Tr } from "@probo/ui";
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
    <>
      {documents.length > 0
        ? (
            <Card>
              <table className="w-full table-fixed">
                <Thead>
                  <Tr>
                    <Th className="text-left">{__("Name")}</Th>
                    <Th className="w-48 text-left">{__("Type")}</Th>
                    <Th className="w-36 text-left">{__("Classification")}</Th>
                    <Th className="w-40 text-left">{__("Last update")}</Th>
                    <Th className="w-32 text-left">{__("Signed")}</Th>
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
    </>
  );
}
