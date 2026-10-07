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

import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { useDeleteConnectorMutation } from "#/__generated__/core/useDeleteConnectorMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const deleteConnectorMutation = graphql`
  mutation useDeleteConnectorMutation($input: DeleteConnectorInput!) {
    deleteConnector(input: $input) {
      deletedConnectorId
    }
  }
`;

export function useDeleteConnector() {
  const { t } = useTranslation("organizations/settings/integrations");
  const organizationId = useOrganizationId();
  const [commit, isDeleting] = useMutation<useDeleteConnectorMutation>(
    deleteConnectorMutation,
    {
      successMessage: t("detailsPage.messages.deleted"),
      errorToast: t("detailsPage.errors.delete"),
    },
  );

  async function deleteConnector(connectorId: string) {
    await commit({
      variables: { input: { connectorId } },
      // Organization.connectors is a plain list, so there is no edge for
      // @deleteEdge to remove.
      updater: (store) => {
        const deletedId = store
          .getRootField("deleteConnector")
          ?.getValue("deletedConnectorId");

        if (typeof deletedId !== "string") {
          return;
        }

        const organization = store.get(organizationId);
        const connectors = organization?.getLinkedRecords("connectors");

        if (organization && connectors) {
          organization.setLinkedRecords(
            connectors.filter(
              connector => connector?.getDataID() !== deletedId,
            ),
            "connectors",
          );
        }

        store.delete(deletedId);
      },
    });
  }

  return [deleteConnector, isDeleting] as const;
}
