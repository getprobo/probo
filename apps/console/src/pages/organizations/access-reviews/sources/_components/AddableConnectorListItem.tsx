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

import { Toast } from "@base-ui/react/toast";
import { type ReactNode, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import type { AddableConnectorListItem_connector$key } from "#/__generated__/core/AddableConnectorListItem_connector.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import {
  createAccessReviewSourcesMutation,
  prependCreatedSourceEdges,
} from "../../dialogs/accessReviewSourceMutations";

import {
  ConnectorAccountPages,
} from "./AddableConnectorAccounts";
import {
  type AddableConnectorCard,
  ResolvedAddableConnectorCard,
} from "./ResolvedAddableConnectorCard";

const fragment = graphql`
  fragment AddableConnectorListItem_connector on Connector @relay(plural: true) {
    ...ResolvedAddableConnectorCard_connector
    ...AddableConnectorAccounts_connector
  }
`;

interface AddableConnectorListItemProps {
  connectorKeys: AddableConnectorListItem_connector$key;
  connectionId: string;
  normalizedSearch: string;
  children: (card: AddableConnectorCard | null) => ReactNode;
}

export function AddableConnectorListItem({
  connectorKeys,
  connectionId,
  normalizedSearch,
  children,
}: AddableConnectorListItemProps) {
  const { t } = useTranslation();
  const toast = Toast.useToastManager();
  const organizationId = useOrganizationId();
  const connectors = useFragment(fragment, connectorKeys);
  const [isAdding, setIsAdding] = useState(false);
  const [createAccessReviewSources, isCreating]
    = useMutation<accessReviewSourceMutationsCreateMutation>(
      createAccessReviewSourcesMutation,
    );
  const busy = isAdding || isCreating;

  async function addSources(accounts: { id: string; name: string }[]) {
    if (busy || accounts.length === 0) {
      return;
    }

    setIsAdding(true);
    try {
      await createAccessReviewSources({
        variables: {
          input: {
            organizationId,
            sources: accounts.map(account => ({
              connectorAccountId: account.id,
              name: account.name,
              connectorId: null,
              csvData: null,
            })),
          },
        },
        updater: (store) => {
          prependCreatedSourceEdges(store, connectionId);
        },
      }, {
        errorToast: t("accessReviewSourcesPage.errors.create"),
      });
      toast.add({
        title: t("accessReviewSourcesPage.messages.created"),
        type: "success",
      });
    } catch {
      // The mutation hook already reported the error.
    } finally {
      setIsAdding(false);
    }
  }

  return (
    <ConnectorAccountPages connectorKeys={connectors}>
      {pages => (
        <ResolvedAddableConnectorCard
          pages={pages}
          connectorKeys={connectors}
          normalizedSearch={normalizedSearch}
          busy={busy}
          onAdd={(accounts) => {
            void addSources(accounts);
          }}
        >
          {children}
        </ResolvedAddableConnectorCard>
      )}
    </ConnectorAccountPages>
  );
}
