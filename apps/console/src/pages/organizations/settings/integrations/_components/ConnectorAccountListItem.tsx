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

import { PlugsIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import { ThirdPartyLogo } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { ConnectionHandler, graphql, useFragment } from "react-relay";

import type { ConnectorAccountListItem_account$key } from "#/__generated__/core/ConnectorAccountListItem_account.graphql";
import type { ConnectorAccountListItem_connector$key } from "#/__generated__/core/ConnectorAccountListItem_connector.graphql";
import type { ConnectorAccountListItemDisconnectMutation } from "#/__generated__/core/ConnectorAccountListItemDisconnectMutation.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import { useMutation } from "#/lib/relay/useMutation";

import { connectorCard } from "../variants";

const connectorAccountListItemFragment = graphql`
  fragment ConnectorAccountListItem_account on ConnectorAccount {
    id
    name
    externalAccountId
    createdAt
  }
`;

const connectorAccountListItemConnectorFragment = graphql`
  fragment ConnectorAccountListItem_connector on Connector {
    id
    provider
    canDelete: permission(action: "core:connector:delete")
    initialAccountExternalId
  }
`;

const disconnectConnectorAccountMutation = graphql`
  mutation ConnectorAccountListItemDisconnectMutation(
    $input: DisableConnectorAccountInput!
  ) {
    disableConnectorAccount(input: $input) {
      disabledConnectorAccountId
    }
  }
`;

interface ConnectorAccountListItemProps {
  accountKey: ConnectorAccountListItem_account$key;
  connectorKey: ConnectorAccountListItem_connector$key;
  unknown: boolean;
  onDisconnected: (externalAccountId: string) => void;
}

export function ConnectorAccountListItem({
  accountKey,
  connectorKey,
  unknown,
  onDisconnected,
}: ConnectorAccountListItemProps) {
  const { t, i18n } = useTranslation("organizations/settings/integrations");
  const account = useFragment(connectorAccountListItemFragment, accountKey);
  const connector = useFragment(connectorAccountListItemConnectorFragment, connectorKey);
  const canDisconnect = connector.canDelete
    && account.externalAccountId !== connector.initialAccountExternalId;
  const { identity, name, title } = connectorCard();
  const [disconnectOpen, setDisconnectOpen] = useState(false);
  const [disconnectAccount, isDisconnecting]
    = useMutation<ConnectorAccountListItemDisconnectMutation>(
      disconnectConnectorAccountMutation,
      {
        successMessage: t("detailsPage.messages.accountDisconnected"),
        errorToast: t("detailsPage.errors.disable"),
      },
    );

  function handleDisconnect() {
    void disconnectAccount({
      variables: { input: { connectorAccountId: account.id } },
      updater: (store) => {
        const deletedId = store
          .getRootField("disableConnectorAccount")
          ?.getValue("disabledConnectorAccountId");
        if (typeof deletedId !== "string") {
          return;
        }

        const connection = store.get(
          ConnectionHandler.getConnectionID(
            connector.id,
            "ConnectorAccountsDrawer_accounts",
          ),
        );
        if (connection != null) {
          ConnectionHandler.deleteNode(connection, deletedId);
          const totalCount: unknown = connection.getValue("totalCount");
          if (typeof totalCount === "number" && totalCount > 0) {
            connection.setValue(totalCount - 1, "totalCount");
          }
        }

        store.delete(deletedId);
      },
    }).then(
      () => {
        onDisconnected(account.externalAccountId);
        setDisconnectOpen(false);
      },
      () => {
        setDisconnectOpen(false);
      },
    );
  }

  return (
    <>
      <TonedCard
        tone={unknown ? "sand" : "green"}
        size={2}
        icon={(
          <ThirdPartyLogo thirdParty={connector.provider} />
        )}
        lead={(
          <div className={identity()}>
            <div className={name()}>
              <Heading level={2} size={3} weight="medium" highContrast className={title()}>
                {account.name}
              </Heading>
              <Text size={1} color="faint" className="shrink-0 font-mono">
                #
                {account.externalAccountId}
              </Text>
            </div>
            <Badge variant="soft" color={unknown ? "neutral" : "green"} size={1}>
              {t(unknown ? "detailsPage.accounts.unknown" : "detailsPage.status.CONNECTED")}
            </Badge>
          </div>
        )}
        control={canDisconnect
          ? (
              <IconButton
                variant="ghost"
                color="red"
                size={1}
                aria-label={t("detailsPage.accounts.disconnect")}
                onClick={() => setDisconnectOpen(true)}
              >
                <PlugsIcon />
              </IconButton>
            )
          : undefined}
      >
        <Text size={1} color="faint">
          <time dateTime={account.createdAt}>
            {t("listPage.created", {
              date: dateTimeFormat(i18n.language, account.createdAt),
            })}
          </time>
        </Text>
      </TonedCard>
      {canDisconnect && (
        <Dialog open={disconnectOpen} onOpenChange={setDisconnectOpen}>
          <DialogPopup>
            <DialogHeader>
              <DialogTitle>
                {t("detailsPage.accounts.disconnectTitle")}
              </DialogTitle>
              <DialogDescription>
                {t("detailsPage.accounts.disconnectDescription", {
                  name: account.name,
                })}
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose
                render={(
                  <Button variant="soft" color="neutral">
                    {t("detailsPage.actions.cancel")}
                  </Button>
                )}
              />
              <Button
                type="button"
                variant="solid"
                color="red"
                iconStart={<PlugsIcon />}
                loading={isDisconnecting}
                onClick={handleDisconnect}
              >
                {t("detailsPage.accounts.disconnectConfirm")}
              </Button>
            </DialogFooter>
          </DialogPopup>
        </Dialog>
      )}
    </>
  );
}
