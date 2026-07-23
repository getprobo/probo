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

import { dateTimeFormat } from "@probo/i18n";
import {
  Badge,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  IconSettingsGear2,
  IconWarning,
  Input,
  Microsoft,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";

import type { Microsoft365ConnectorDeleteMutation } from "#/__generated__/iam/Microsoft365ConnectorDeleteMutation.graphql";
import type { Microsoft365ConnectorFragment$key } from "#/__generated__/iam/Microsoft365ConnectorFragment.graphql";
import type { Microsoft365ConnectorUpdateSCIMBridgeMutation } from "#/__generated__/iam/Microsoft365ConnectorUpdateSCIMBridgeMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const microsoft365ConnectorFragment = graphql`
  fragment Microsoft365ConnectorFragment on SCIMConfiguration {
    id
    bridge {
      id
      type
      state
      syncError
      excludedUserNames
      connector {
        id
        createdAt
      }
    }
  }
`;

const deleteSCIMConfigurationMutation = graphql`
  mutation Microsoft365ConnectorDeleteMutation(
    $input: DeleteSCIMConfigurationInput!
  ) {
    deleteSCIMConfiguration(input: $input) {
      deletedScimConfigurationId @deleteRecord
    }
  }
`;

const updateSCIMBridgeMutation = graphql`
  mutation Microsoft365ConnectorUpdateSCIMBridgeMutation(
    $input: UpdateSCIMBridgeInput!
  ) {
    updateSCIMBridge(input: $input) {
      scimBridge {
        id
        excludedUserNames
      }
    }
  }
`;

export function Microsoft365Connector(props: {
  fKey: Microsoft365ConnectorFragment$key | null;
  oauth2Scopes: readonly string[];
}) {
  const { fKey, oauth2Scopes } = props;
  const data = useFragment<Microsoft365ConnectorFragment$key>(microsoft365ConnectorFragment, fKey);
  const bridge = data?.bridge?.type === "MICROSOFT_365" ? data.bridge : null;
  const connector = bridge?.connector;
  const scimConfigurationId = data?.id;
  const bridgeId = bridge?.id;
  const organizationId = useOrganizationId();
  const { t, i18n } = useTranslation();
  const bridgeState = bridge?.state ?? null;
  const latestBridgeError = bridge?.syncError ?? null;
  const isBridgeFailed = bridgeState === "FAILED";
  const isBridgeDisabled = bridgeState === "DISABLED";
  const hasBridgeError = isBridgeFailed || isBridgeDisabled;
  const isBridgePending = bridgeState === "PENDING";
  const bridgeStatusBadgeVariant = hasBridgeError
    ? "danger"
    : isBridgePending
      ? "warning"
      : "success";
  const bridgeStatusLabel = isBridgeDisabled
    ? t("microsoft365Connector.status.disabled")
    : isBridgeFailed
      ? t("microsoft365Connector.status.error")
      : isBridgePending
        ? t("microsoft365Connector.status.syncing")
        : t("microsoft365Connector.status.connected");
  const bridgeErrorMessage = latestBridgeError
    ?? t("microsoft365Connector.errors.bridgeSync");
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const excludedUserNamesDialogRef = useDialogRef();

  const [newUser, setNewUser] = useState("");

  const [deleteSCIMConfiguration, isDeleting]
    = useMutation<Microsoft365ConnectorDeleteMutation>(
      deleteSCIMConfigurationMutation,
    );

  const [updateSCIMBridge, isUpdating]
    = useMutation<Microsoft365ConnectorUpdateSCIMBridgeMutation>(
      updateSCIMBridgeMutation,
    );

  const handleConnect = () => {
    const baseUrl = import.meta.env.VITE_API_URL || window.location.origin;
    const url = new URL("/api/console/v1/connectors/initiate", baseUrl);
    url.searchParams.append("organization_id", organizationId);
    url.searchParams.append("provider", "MICROSOFT_365");
    for (const scope of oauth2Scopes) {
      url.searchParams.append("scope", scope);
    }
    const continueUrl = `/organizations/${organizationId}/settings/scim`;
    url.searchParams.append("continue", continueUrl);
    window.location.href = url.toString();
  };

  const handleDisconnect = () => {
    if (!connector || !scimConfigurationId) return;

    void deleteSCIMConfiguration({
      variables: {
        input: {
          organizationId: organizationId,
          scimConfigurationId: scimConfigurationId,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: t("common.error"),
            description: errors.map(e => e.message).join(", "),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("common.success"),
          description: t("microsoft365Connector.messages.disconnected"),
          variant: "success",
        });
        dialogRef.current?.close();
      },
      onError(error) {
        toast({
          title: t("common.error"),
          description: error.message,
          variant: "error",
        });
      },
    });
  };

  const currentExcludedUserNames = bridge?.excludedUserNames ? [...bridge.excludedUserNames] : [];

  const saveExcludedUserNames = (newList: string[]) => {
    if (!bridgeId) return;

    void updateSCIMBridge({
      variables: {
        input: {
          organizationId: organizationId,
          scimBridgeId: bridgeId,
          excludedUserNames: newList,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: t("common.error"),
            description: errors.map(e => e.message).join(", "),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("common.success"),
          description: t("microsoft365Connector.messages.excludedUsersUpdated"),
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
  };

  const handleAddUser = () => {
    const user = newUser.trim().toLowerCase();
    if (user && !currentExcludedUserNames.includes(user)) {
      saveExcludedUserNames([...currentExcludedUserNames, user]);
      setNewUser("");
    }
  };

  const handleRemoveUser = (user: string) => {
    saveExcludedUserNames(currentExcludedUserNames.filter(e => e !== user));
  };

  if (!connector) {
    return (
      <Card padded className="flex items-center gap-3">
        <div className="w-10 h-10 flex items-center justify-center bg-subtle rounded">
          <Microsoft className="w-6 h-6" />
        </div>
        <div className="mr-auto">
          <h3 className="font-medium">{t("microsoft365Connector.name")}</h3>
          <p className="text-sm text-txt-secondary">
            {t("microsoft365Connector.connectDescription")}
          </p>
        </div>
        <Button variant="secondary" onClick={handleConnect}>
          {t("microsoft365Connector.actions.connect")}
        </Button>
      </Card>
    );
  }

  return (
    <Card padded className="space-y-3">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 flex items-center justify-center bg-subtle rounded">
          <Microsoft className="w-6 h-6" />
        </div>
        <div className="mr-auto">
          <h3 className="font-medium">{t("microsoft365Connector.name")}</h3>
          <p className="text-sm text-txt-secondary">
            {t("microsoft365Connector.connectedOn", { date: dateTimeFormat(i18n.language, connector.createdAt) })}
          </p>
        </div>
        <Badge variant={bridgeStatusBadgeVariant} size="md">
          {bridgeStatusLabel}
        </Badge>
        <Dialog
          ref={excludedUserNamesDialogRef}
          trigger={(
            <Button variant="secondary">
              <IconSettingsGear2 size={16} />
              {t("microsoft365Connector.actions.settings")}
            </Button>
          )}
          title={t("microsoft365Connector.settings.title")}
          className="max-w-lg"
        >
          <DialogContent padded className="space-y-6">
            <div className="space-y-4">
              <div>
                <h4 className="text-sm font-medium">{t("microsoft365Connector.settings.excludedUserNames")}</h4>
                <p className="text-sm text-txt-secondary mt-1">
                  {t("microsoft365Connector.settings.excludedUserNamesDescription")}
                </p>
              </div>
              <div className="flex gap-2">
                <Input
                  type="text"
                  value={newUser}
                  onChange={e => setNewUser(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault();
                      if (isUpdating) return;
                      handleAddUser();
                    }
                  }}
                  placeholder="user@example.com"
                  className="flex-1"
                />
                <Button onClick={handleAddUser} variant="secondary" disabled={isUpdating}>
                  {t("microsoft365Connector.actions.add")}
                </Button>
              </div>

              {currentExcludedUserNames.length > 0 && (
                <div className="space-y-2">
                  {currentExcludedUserNames.map((user: string) => (
                    <div
                      key={user}
                      className="flex items-center justify-between p-2 bg-subtle rounded"
                    >
                      <span className="text-sm">{user}</span>
                      <Button
                        variant="quaternary"
                        onClick={() => handleRemoveUser(user)}
                        disabled={isUpdating}
                      >
                        {t("microsoft365Connector.actions.remove")}
                      </Button>
                    </div>
                  ))}
                </div>
              )}

              {currentExcludedUserNames.length === 0 && (
                <p className="text-sm text-txt-secondary text-center py-4">
                  {t("microsoft365Connector.settings.noExcludedUserNames")}
                </p>
              )}
            </div>
          </DialogContent>
        </Dialog>
        <Dialog
          ref={dialogRef}
          trigger={(
            <Button variant="danger">
              {t("microsoft365Connector.actions.disconnect")}
            </Button>
          )}
          title={t("microsoft365Connector.disconnect.title")}
          className="max-w-lg"
        >
          <DialogContent padded className="space-y-4">
            <p className="text-txt-secondary text-sm">
              {t("microsoft365Connector.disconnect.description")}
            </p>
            <p className="text-red-600 text-sm font-medium">
              {t("microsoft365Connector.disconnect.warning")}
            </p>
          </DialogContent>
          <DialogFooter>
            <Button
              variant="danger"
              onClick={handleDisconnect}
              disabled={isDeleting}
            >
              {isDeleting
                ? t("microsoft365Connector.actions.disconnecting")
                : t("microsoft365Connector.actions.disconnect")}
            </Button>
          </DialogFooter>
        </Dialog>
      </div>

      {hasBridgeError && (
        <div className="flex items-start gap-2 rounded-lg bg-bg-warning/10 border border-border-warning p-3">
          <IconWarning size={16} className="text-txt-danger shrink-0 mt-0.5" />
          <div className="space-y-1">
            <p className="text-sm font-medium text-txt-danger">
              {isBridgeDisabled
                ? t("microsoft365Connector.errors.bridgeDisabled")
                : t("microsoft365Connector.errors.bridgeFailed")}
            </p>
            <p className="text-sm text-txt-secondary whitespace-pre-wrap break-all">
              {bridgeErrorMessage}
            </p>
          </div>
        </div>
      )}
    </Card>
  );
}
