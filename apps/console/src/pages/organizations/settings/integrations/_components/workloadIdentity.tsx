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

import { CopyIcon } from "@phosphor-icons/react";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { workloadIdentityCreateMutation } from "#/__generated__/core/workloadIdentityCreateMutation.graphql";
import type { workloadIdentityDeleteMutation } from "#/__generated__/core/workloadIdentityDeleteMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { createdConnectorState } from "../_lib/discoveredAccounts";
import { connectorDetailsPath } from "../_lib/integrationPath";

import type { ConnectDestination } from "./ConnectForm";

const createWorkloadIdentityConnectorMutation = graphql`
  mutation workloadIdentityCreateMutation($input: CreateWorkloadIdentityConnectorInput!) {
    createWorkloadIdentityConnector(input: $input) {
      connector {
        id
        connectionStatus
        discoveredAccounts {
          externalAccountId
        }
      }
    }
  }
`;

const deleteConnectorMutation = graphql`
  mutation workloadIdentityDeleteMutation($input: DeleteConnectorInput!) {
    deleteConnector(input: $input) {
      deletedConnectorId
    }
  }
`;

const setupFields = ["issuer", "audience", "subject"] as const;

export function TerraformInstallButton({ snippet }: { snippet: string }) {
  const { t } = useTranslation("organizations/settings/integrations");
  const copyValue = useCopyValue();

  return (
    <Button
      type="button"
      variant="soft"
      className="shrink-0"
      onClick={() => copyValue(
        snippet,
        t("marketplacePage.workloadIdentity.messages.copiedTerraform"),
        t("marketplacePage.workloadIdentity.messages.copyFailed"),
      )}
    >
      {t("marketplacePage.workloadIdentity.actions.installViaTerraform")}
    </Button>
  );
}

export function useCopyValue() {
  const { toast } = useToast();

  return (value: string, title: string, failure: string) => {
    const onCopyFailure = () => {
      toast({ title: failure, description: failure, variant: "error" });
    };
    if (!navigator.clipboard?.writeText) {
      onCopyFailure();
      return;
    }
    navigator.clipboard.writeText(value).then(
      () => toast({ title, description: title, variant: "success" }),
      onCopyFailure,
    );
  };
}

export function setupRows(
  t: (key: string) => string,
  pageKey: string,
  setup: { issuer: string; audience: string; subject: string },
  copyValue: (value: string, title: string, failure: string) => void,
) {
  const copiedKey = {
    issuer: "copiedIssuer",
    audience: "copiedAudience",
    subject: "copiedSubject",
  } as const;

  return setupFields.map(field => ({
    label: t(`${pageKey}.fields.${field}`),
    value: setup[field],
    copyLabel: t(`${pageKey}.actions.copy`),
    onCopy: () => copyValue(
      setup[field],
      t(`${pageKey}.messages.${copiedKey[field]}`),
      t(`${pageKey}.messages.copyFailed`),
    ),
  }));
}

export function SetupValues({
  rows,
}: {
  rows: Array<{ label: string; value: string; copyLabel: string; onCopy: () => void }>;
}) {
  return (
    <div className="flex flex-col gap-2 rounded-2 bg-sand-3 p-3">
      {rows.map(row => (
        <div key={row.label} className="flex flex-col gap-1">
          <Text size={1} color="faint">{row.label}</Text>
          <div className="flex min-w-0 items-center gap-1">
            <Code size={2} className="min-w-0 flex-1 break-all">{row.value}</Code>
            <IconButton
              type="button"
              size={1}
              variant="soft"
              color="neutral"
              aria-label={row.copyLabel}
              onClick={row.onCopy}
            >
              <CopyIcon />
            </IconButton>
          </div>
        </div>
      ))}
    </div>
  );
}

export function useFinishWorkloadIdentity(organizationId: string) {
  const { toast } = useToast();
  const [createConnector] = useMutation<workloadIdentityCreateMutation>(createWorkloadIdentityConnectorMutation);
  const [deleteConnector] = useMutation<workloadIdentityDeleteMutation>(deleteConnectorMutation);

  return async function finish(
    input: workloadIdentityCreateMutation["variables"]["input"],
    errors: { create: string; disconnected: string; delete: string; errorTitle: string },
  ): Promise<ConnectDestination | null> {
    const created = await createConnector({
      variables: { input },
    }, { errorToast: errors.create });
    const connector = created.createWorkloadIdentityConnector.connector;
    if (connector.connectionStatus !== "CONNECTED") {
      toast({
        title: errors.errorTitle,
        description: errors.disconnected,
        variant: "error",
      });
      await deleteConnector({
        variables: { input: { connectorId: connector.id } },
      }, { errorToast: errors.delete });
      return null;
    }

    return {
      to: connectorDetailsPath(organizationId, input.provider),
      state: connector.discoveredAccounts.length === 0
        ? undefined
        : createdConnectorState(connector.id),
    };
  };
}
