// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
import { useTranslate } from "@probo/i18n";
import { Card, Tbody, Th, Thead, Tr } from "@probo/ui";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { EmployeeApprovalsPageQuery } from "#/__generated__/core/EmployeeApprovalsPageQuery.graphql";

import { ApprovableDocumentRow } from "./_components/ApprovableDocumentRow";

export const employeeApprovalsPageQuery = graphql`
  query EmployeeApprovalsPageQuery($organizationId: ID!) {
    viewer @required(action: THROW) {
      approvableDocuments(
        organizationId: $organizationId
        first: 1000
        orderBy: { field: UPDATED_AT, direction: DESC }
      ) @required(action: THROW) {
        edges @required(action: THROW) {
          node @required(action: THROW) {
            id
            ...ApprovableDocumentRowFragment
          }
        }
      }
    }
  }
`;

export function EmployeeApprovalsPage(props: {
  queryRef: PreloadedQuery<EmployeeApprovalsPageQuery>;
}) {
  const { queryRef } = props;
  const { __ } = useTranslate();

  const {
    viewer: { approvableDocuments },
  } = usePreloadedQuery<EmployeeApprovalsPageQuery>(
    employeeApprovalsPageQuery,
    queryRef,
  );

  const documents = approvableDocuments.edges.map(edge => edge.node);

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
                    <Th className="w-32 text-left">{__("Approved")}</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {documents.map(document => (
                    <ApprovableDocumentRow
                      key={document.id}
                      fKey={document}
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
                  {__("No documents have been requested for your approval.")}
                </p>
              </div>
            </Card>
          )}
    </>
  );
}
