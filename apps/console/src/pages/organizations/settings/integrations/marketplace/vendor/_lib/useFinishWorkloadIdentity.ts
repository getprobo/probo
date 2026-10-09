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

import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { graphql } from "react-relay";

import type { useFinishWorkloadIdentityCreateMutation } from "#/__generated__/core/useFinishWorkloadIdentityCreateMutation.graphql";
import type { useFinishWorkloadIdentityDeleteMutation } from "#/__generated__/core/useFinishWorkloadIdentityDeleteMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { createdConnectorState } from "../../../_lib/discoveredAccounts";
import { connectorDetailsPath } from "../../../_lib/integrationPath";
import type { ConnectDestination } from "../_components/ConnectForm";

const createWorkloadIdentityConnectorMutation = graphql`
  mutation useFinishWorkloadIdentityCreateMutation($input: CreateWorkloadIdentityConnectorInput!) {
    createWorkloadIdentityConnector(input: $input) {
      connector {
        id
        connectionStatus
      }
    }
  }
`;

const deleteConnectorMutation = graphql`
  mutation useFinishWorkloadIdentityDeleteMutation($input: DeleteConnectorInput!) {
    deleteConnector(input: $input) {
      deletedConnectorId
    }
  }
`;

export function useFinishWorkloadIdentity() {
  const organizationId = useOrganizationId();
  const toast = useToast();
  const [createConnector] = useMutation<useFinishWorkloadIdentityCreateMutation>(
    createWorkloadIdentityConnectorMutation,
  );
  const [deleteConnector] = useMutation<useFinishWorkloadIdentityDeleteMutation>(
    deleteConnectorMutation,
  );

  return async function finish(
    input: useFinishWorkloadIdentityCreateMutation["variables"]["input"],
    errors: { create: string; disconnected: string; delete: string; errorTitle: string },
  ): Promise<ConnectDestination | null> {
    const created = await createConnector({
      variables: { input },
    }, { errorToast: errors.create });
    const connector = created.createWorkloadIdentityConnector.connector;
    if (connector.connectionStatus !== "CONNECTED") {
      toast.add({
        title: errors.errorTitle,
        description: errors.disconnected,
        type: "error",
      });
      await deleteConnector({
        variables: { input: { connectorId: connector.id } },
      }, { errorToast: errors.delete });
      return null;
    }

    return {
      to: connectorDetailsPath(organizationId, input.provider),
      state: createdConnectorState(connector.id),
    };
  };
}
