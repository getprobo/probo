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

import { GearIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import {
  Button as LegacyButton,
  Dialog,
  DialogContent,
  DialogFooter,
  Input,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { GoogleLogo } from "@probo/ui/src/v2/GoogleLogo/GoogleLogo";
import { MicrosoftLogo } from "@probo/ui/src/v2/MicrosoftLogo/MicrosoftLogo";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";

import type { SCIMProviderCard_deleteMutation } from "#/__generated__/iam/SCIMProviderCard_deleteMutation.graphql";
import type { SCIMProviderCard_scimConfiguration$key } from "#/__generated__/iam/SCIMProviderCard_scimConfiguration.graphql";
import type { SCIMProviderCard_updateMutation } from "#/__generated__/iam/SCIMProviderCard_updateMutation.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import type { TonedCardTone } from "#/components/TonedCard/variants";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { scimProviderCard } from "../variants";

import { ReactivateSCIMBridgeButton } from "./ReactivateSCIMBridgeButton";

const scimProviderCardFragment = graphql`
  fragment SCIMProviderCard_scimConfiguration on SCIMConfiguration {
    id
    canDelete: permission(action: "iam:scim-configuration:delete")
    bridge {
      id
      type
      state
      syncError
      excludedUserNames
      createdAt
      canUpdate: permission(action: "iam:scim-bridge:update")
      connector {
        createdAt
      }
      ...ReactivateSCIMBridgeButtonFragment
    }
  }
`;

const deleteSCIMConfigurationMutation = graphql`
  mutation SCIMProviderCard_deleteMutation(
    $input: DeleteSCIMConfigurationInput!
  ) {
    deleteSCIMConfiguration(input: $input) {
      deletedScimConfigurationId @deleteRecord
    }
  }
`;

const updateSCIMBridgeMutation = graphql`
  mutation SCIMProviderCard_updateMutation($input: UpdateSCIMBridgeInput!) {
    updateSCIMBridge(input: $input) {
      scimBridge {
        id
        excludedUserNames
      }
    }
  }
`;

type ProviderCopy
  = "googleWorkspaceConnector"
    | "microsoft365Connector";

type BridgeStatus = "connected" | "syncing" | "error" | "disabled";

const statusTone: Record<BridgeStatus, TonedCardTone> = {
  connected: "green",
  syncing: "amber",
  error: "red",
  disabled: "sand",
};

const statusBadgeColor: Record<
  BridgeStatus,
  "green" | "amber" | "red" | "neutral"
> = {
  connected: "green",
  syncing: "amber",
  error: "red",
  disabled: "neutral",
};

function bridgeStatus(state: string): BridgeStatus {
  if (state === "DISABLED") {
    return "disabled";
  }
  if (state === "FAILED") {
    return "error";
  }
  if (state === "PENDING" || state === "SYNCING") {
    return "syncing";
  }
  return "connected";
}

export interface SCIMProviderCardProps {
  scimConfigurationKey: SCIMProviderCard_scimConfiguration$key;
}

export function SCIMProviderCard({ scimConfigurationKey }: SCIMProviderCardProps) {
  const data = useFragment(scimProviderCardFragment, scimConfigurationKey);
  const bridge = data.bridge;
  const { t, i18n } = useTranslation();
  const organizationId = useOrganizationId();
  const { toast } = useToast();
  const settingsRef = useDialogRef();
  const disconnectRef = useDialogRef();
  const [newUser, setNewUser] = useState("");
  const { actions, body, callout, hint } = scimProviderCard();

  const [deleteSCIMConfiguration, isDeleting]
    = useMutation<SCIMProviderCard_deleteMutation>(deleteSCIMConfigurationMutation);
  const [updateSCIMBridge, isUpdating]
    = useMutation<SCIMProviderCard_updateMutation>(updateSCIMBridgeMutation);

  if (bridge == null) {
    return null;
  }

  const copy: ProviderCopy = bridge.type === "MICROSOFT_365"
    ? "microsoft365Connector"
    : "googleWorkspaceConnector";
  const status = bridgeStatus(bridge.state);
  const tone = statusTone[status];
  const hasError = status === "error" || status === "disabled";
  const connectedAt = bridge.connector?.createdAt ?? bridge.createdAt;
  const excludedUserNames = [...bridge.excludedUserNames];
  const bridgeId = bridge.id;

  function saveExcludedUserNames(next: string[]) {
    void updateSCIMBridge({
      variables: {
        input: {
          organizationId,
          scimBridgeId: bridgeId,
          excludedUserNames: next,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: t("common.error"),
            description: errors.map(error => error.message).join(", "),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("common.success"),
          description: t(`${copy}.messages.excludedUsersUpdated`),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: t("common.error"),
          description: error.message,
          variant: "error",
        });
      },
    });
  }

  function handleAddUser() {
    const user = newUser.trim().toLowerCase();
    if (user === "" || excludedUserNames.includes(user)) {
      return;
    }
    saveExcludedUserNames([...excludedUserNames, user]);
    setNewUser("");
  }

  function handleDisconnect() {
    void deleteSCIMConfiguration({
      variables: {
        input: {
          organizationId,
          scimConfigurationId: data.id,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: t("common.error"),
            description: errors.map(error => error.message).join(", "),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("common.success"),
          description: t(`${copy}.messages.disconnected`),
          variant: "success",
        });
        disconnectRef.current?.close();
      },
      onError(error) {
        toast({
          title: t("common.error"),
          description: error.message,
          variant: "error",
        });
      },
    });
  }

  return (
    <TonedCard
      tone={tone}
      icon={bridge.type === "MICROSOFT_365"
        ? <MicrosoftLogo className="size-6" />
        : <GoogleLogo className="size-6" />}
      lead={(
        <Text size={3} weight="medium" color={tone === "sand" ? "neutral" : tone}>
          {t(`${copy}.name`)}
        </Text>
      )}
      control={(
        <Badge size={2} variant="soft" color={statusBadgeColor[status]}>
          {t(`${copy}.status.${status}`)}
        </Badge>
      )}
    >
      <div className={body()}>
        <Text size={2} color="neutral">
          {t(`${copy}.connectedOn`, {
            date: dateTimeFormat(i18n.language, connectedAt),
          })}
        </Text>
        {hasError && (
          <div className={callout()}>
            <Text size={2} weight="medium" highContrast>
              {status === "disabled"
                ? t(`${copy}.errors.bridgeDisabled`)
                : t(`${copy}.errors.bridgeFailed`)}
            </Text>
            <Text size={2} color="neutral">
              {bridge.syncError ?? t(`${copy}.errors.bridgeSync`)}
            </Text>
            <ReactivateSCIMBridgeButton fKey={bridge} />
          </div>
        )}
        <div className={actions()}>
          <Text size={2} color="neutral" className={hint()}>
            {t(`${copy}.excludedCount`, { count: excludedUserNames.length })}
          </Text>
          {bridge.canUpdate && (
            <Button
              size={2}
              variant="surface"
              color="neutral"
              iconStart={<GearIcon />}
              onClick={() => settingsRef.current?.open()}
            >
              {t(`${copy}.actions.settings`)}
            </Button>
          )}
          {data.canDelete && (
            <Button
              size={2}
              variant="solid"
              color="red"
              onClick={() => disconnectRef.current?.open()}
            >
              {t(`${copy}.actions.disconnect`)}
            </Button>
          )}
        </div>
      </div>

      <Dialog
        ref={settingsRef}
        title={t(`${copy}.settings.title`)}
        className="max-w-lg"
      >
        <DialogContent padded className="space-y-6">
          <div className="space-y-4">
            <div>
              <h4 className="text-sm font-medium">{t(`${copy}.settings.excludedUserNames`)}</h4>
              <p className="text-sm text-txt-secondary mt-1">
                {t(`${copy}.settings.excludedUserNamesDescription`)}
              </p>
            </div>
            <div className="flex gap-2">
              <Input
                type="text"
                value={newUser}
                onChange={event => setNewUser(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter") {
                    event.preventDefault();
                    if (isUpdating) {
                      return;
                    }
                    handleAddUser();
                  }
                }}
                placeholder="user@example.com"
                className="flex-1"
              />
              <LegacyButton
                onClick={handleAddUser}
                variant="secondary"
                disabled={isUpdating}
              >
                {t(`${copy}.actions.add`)}
              </LegacyButton>
            </div>
            {excludedUserNames.length === 0
              ? (
                  <p className="text-sm text-txt-secondary text-center py-4">
                    {t(`${copy}.settings.noExcludedUserNames`)}
                  </p>
                )
              : (
                  <div className="space-y-2">
                    {excludedUserNames.map(user => (
                      <div
                        key={user}
                        className="flex items-center justify-between p-2 bg-subtle rounded"
                      >
                        <span className="text-sm">{user}</span>
                        <LegacyButton
                          variant="quaternary"
                          onClick={() => {
                            saveExcludedUserNames(
                              excludedUserNames.filter(name => name !== user),
                            );
                          }}
                          disabled={isUpdating}
                        >
                          {t(`${copy}.actions.remove`)}
                        </LegacyButton>
                      </div>
                    ))}
                  </div>
                )}
          </div>
        </DialogContent>
      </Dialog>

      <Dialog
        ref={disconnectRef}
        title={t(`${copy}.disconnect.title`)}
        className="max-w-lg"
      >
        <DialogContent padded className="space-y-4">
          <p className="text-txt-secondary text-sm">
            {t(`${copy}.disconnect.description`)}
          </p>
          <p className="text-red-600 text-sm font-medium">
            {t(`${copy}.disconnect.warning`)}
          </p>
        </DialogContent>
        <DialogFooter>
          <LegacyButton
            variant="danger"
            onClick={handleDisconnect}
            disabled={isDeleting}
          >
            {isDeleting
              ? t(`${copy}.actions.disconnecting`)
              : t(`${copy}.actions.disconnect`)}
          </LegacyButton>
        </DialogFooter>
      </Dialog>
    </TonedCard>
  );
}
