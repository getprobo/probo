// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { usePageTitle } from "@probo/hooks";
import { Card, Tbody, Th, Thead, Tr } from "@probo/ui";
import { useTranslation } from "react-i18next";
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
        orderBy: { field: UPDATED_AT, direction: DESC }
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
  const { t } = useTranslation();
  const organizationId = useOrganizationId();

  const {
    viewer: { signableDocuments },
  } = usePreloadedQuery<EmployeeDocumentsPageQuery>(
    employeeDocumentsPageQuery,
    queryRef,
  );

  const documents = signableDocuments.edges.map(edge => edge.node);

  usePageTitle(t("employeeDocumentsPage.pageTitle"));

  return (
    <>
      {documents.length > 0
        ? (
            <Card>
              <table className="w-full table-fixed">
                <Thead>
                  <Tr>
                    <Th className="text-left">{t("employeeDocumentsPage.columns.name")}</Th>
                    <Th className="w-48 text-left">{t("employeeDocumentsPage.columns.type")}</Th>
                    <Th className="w-36 text-left">{t("employeeDocumentsPage.columns.classification")}</Th>
                    <Th className="w-40 text-left">{t("employeeDocumentsPage.columns.lastUpdate")}</Th>
                    <Th className="w-32 text-left">{t("employeeDocumentsPage.columns.signed")}</Th>
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
                  {t("employeeDocumentsPage.empty.title")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {t("employeeDocumentsPage.empty.description")}
                </p>
              </div>
            </Card>
          )}
    </>
  );
}
