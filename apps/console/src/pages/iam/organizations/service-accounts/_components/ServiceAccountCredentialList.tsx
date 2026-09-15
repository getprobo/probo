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

import { KeyIcon } from "@phosphor-icons/react";
import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ServiceAccountCredentialList_serviceAccount$key } from "#/__generated__/iam/ServiceAccountCredentialList_serviceAccount.graphql";

import { ServiceAccountCredentialListItem } from "./ServiceAccountCredentialListItem";

const serviceAccountCredentialListFragment = graphql`
  fragment ServiceAccountCredentialList_serviceAccount on ServiceAccount {
    ...ServiceAccountCredentialListItem_serviceAccount
    credentials(
      first: 50
      orderBy: { field: CREATED_AT, direction: DESC }
    ) @connection(key: "ServiceAccountCredentialList_credentials", filters: []) {
      edges {
        node {
          id
          ...ServiceAccountCredentialListItem_credential
        }
      }
    }
  }
`;

interface ServiceAccountCredentialListProps {
  serviceAccountKey: ServiceAccountCredentialList_serviceAccount$key;
}

export function ServiceAccountCredentialList(
  { serviceAccountKey }: ServiceAccountCredentialListProps,
) {
  const { t } = useTranslation("iam/organizations/service-accounts");
  const serviceAccount = useFragment(
    serviceAccountCredentialListFragment,
    serviceAccountKey,
  );
  const credentials = serviceAccount.credentials.edges;

  if (credentials.length === 0) {
    return (
      <div className="flex flex-col items-center gap-2 rounded-3 border border-sand-6 px-4 py-8 text-center">
        <KeyIcon className="size-6 text-sand-10" aria-hidden />
        <Text size={2} weight="medium" highContrast>
          {t("credentials.empty")}
        </Text>
      </div>
    );
  }

  return (
    <Table size={1} variant="surface">
      <TableHeader>
        <TableRow>
          <TableColumnHeaderCell>{t("credentials.columns.name")}</TableColumnHeaderCell>
          <TableColumnHeaderCell>{t("credentials.columns.status")}</TableColumnHeaderCell>
          <TableColumnHeaderCell>{t("credentials.columns.expiresAt")}</TableColumnHeaderCell>
          <TableColumnHeaderCell>{t("credentials.columns.lastUsedAt")}</TableColumnHeaderCell>
          <TableColumnHeaderCell>{t("columns.actions")}</TableColumnHeaderCell>
        </TableRow>
      </TableHeader>
      <TableBody>
        {credentials.map(({ node }) => (
          <ServiceAccountCredentialListItem
            key={node.id}
            credentialKey={node}
            serviceAccountKey={serviceAccount}
          />
        ))}
      </TableBody>
    </Table>
  );
}
