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
  Google,
  IconSettingsGear2,
  IconWarning,
  Input,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";

import type { GoogleWorkspaceConnectorDeleteMutation } from "#/__generated__/iam/GoogleWorkspaceConnectorDeleteMutation.graphql";
import type { GoogleWorkspaceConnectorFragment$key } from "#/__generated__/iam/GoogleWorkspaceConnectorFragment.graphql";
import type { GoogleWorkspaceConnectorUpdateSCIMBridgeMutation } from "#/__generated__/iam/GoogleWorkspaceConnectorUpdateSCIMBridgeMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const googleWorkspaceConnectorFragment = graphql`
  fragment GoogleWorkspaceConnectorFragment on SCIMConfiguration {
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
  mutation GoogleWorkspaceConnectorDeleteMutation(
    $input: DeleteSCIMConfigurationInput!
  ) {
    deleteSCIMConfiguration(input: $input) {
      deletedScimConfigurationId @deleteRecord
    }
  }
`;

const updateSCIMBridgeMutation = graphql`
  mutation GoogleWorkspaceConnectorUpdateSCIMBridgeMutation(
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

export function GoogleWorkspaceConnector(props: {
  fKey: GoogleWorkspaceConnectorFragment$key | null;
  oauth2Scopes: readonly string[];
}) {
  const { fKey, oauth2Scopes } = props;
  const data = useFragment<GoogleWorkspaceConnectorFragment$key>(googleWorkspaceConnectorFragment, fKey);
  const bridge = data?.bridge?.type === "GOOGLE_WORKSPACE" ? data.bridge : null;
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
    ? t("googleWorkspaceConnector.status.disabled")
    : isBridgeFailed
      ? t("googleWorkspaceConnector.status.error")
      : isBridgePending
        ? t("googleWorkspaceConnector.status.syncing")
        : t("googleWorkspaceConnector.status.connected");
  const bridgeErrorMessage = latestBridgeError
    ?? t("googleWorkspaceConnector.errors.bridgeSync");
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const excludedUserNamesDialogRef = useDialogRef();

  const [newUser, setNewUser] = useState("");

  const [deleteSCIMConfiguration, isDeleting]
    = useMutation<GoogleWorkspaceConnectorDeleteMutation>(
      deleteSCIMConfigurationMutation,
    );

  const [updateSCIMBridge, isUpdating]
    = useMutation<GoogleWorkspaceConnectorUpdateSCIMBridgeMutation>(
      updateSCIMBridgeMutation,
    );

  const handleConnect = () => {
    const baseUrl = import.meta.env.VITE_API_URL || window.location.origin;
    const url = new URL("/api/console/v1/connectors/initiate", baseUrl);
    url.searchParams.append("organization_id", organizationId);
    url.searchParams.append("provider", "GOOGLE_WORKSPACE");
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
          description: t("googleWorkspaceConnector.messages.disconnected"),
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
          description: t("googleWorkspaceConnector.messages.excludedUsersUpdated"),
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

  // Not connected state
  if (!connector) {
    return (
      <Card padded className="flex items-center gap-3">
        <div className="w-10 h-10 flex items-center justify-center bg-subtle rounded">
          <Google className="w-6 h-6" />
        </div>
        <div className="mr-auto">
          <h3 className="font-medium">{t("googleWorkspaceConnector.name")}</h3>
          <p className="text-sm text-txt-secondary">
            {t("googleWorkspaceConnector.connectDescription")}
          </p>
        </div>
        <Button variant="secondary" onClick={handleConnect}>
          {t("googleWorkspaceConnector.actions.connect")}
        </Button>
      </Card>
    );
  }

  // Connected state
  return (
    <Card padded className="space-y-3">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 flex items-center justify-center bg-subtle rounded">
          <Google className="w-6 h-6" />
        </div>
        <div className="mr-auto">
          <h3 className="font-medium">{t("googleWorkspaceConnector.name")}</h3>
          <p className="text-sm text-txt-secondary">
            {t("googleWorkspaceConnector.connectedOn", { date: dateTimeFormat(i18n.language, connector.createdAt) })}
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
              {t("googleWorkspaceConnector.actions.settings")}
            </Button>
          )}
          title={t("googleWorkspaceConnector.settings.title")}
          className="max-w-lg"
        >
          <DialogContent padded className="space-y-6">
            <div className="space-y-4">
              <div>
                <h4 className="text-sm font-medium">{t("googleWorkspaceConnector.settings.excludedUserNames")}</h4>
                <p className="text-sm text-txt-secondary mt-1">
                  {t("googleWorkspaceConnector.settings.excludedUserNamesDescription")}
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
                  {t("googleWorkspaceConnector.actions.add")}
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
                        {t("googleWorkspaceConnector.actions.remove")}
                      </Button>
                    </div>
                  ))}
                </div>
              )}

              {currentExcludedUserNames.length === 0 && (
                <p className="text-sm text-txt-secondary text-center py-4">
                  {t("googleWorkspaceConnector.settings.noExcludedUserNames")}
                </p>
              )}
            </div>
          </DialogContent>
        </Dialog>
        <Dialog
          ref={dialogRef}
          trigger={(
            <Button variant="danger">
              {t("googleWorkspaceConnector.actions.disconnect")}
            </Button>
          )}
          title={t("googleWorkspaceConnector.disconnect.title")}
          className="max-w-lg"
        >
          <DialogContent padded className="space-y-4">
            <p className="text-txt-secondary text-sm">
              {t("googleWorkspaceConnector.disconnect.description")}
            </p>
            <p className="text-red-600 text-sm font-medium">
              {t("googleWorkspaceConnector.disconnect.warning")}
            </p>
          </DialogContent>
          <DialogFooter>
            <Button
              variant="danger"
              onClick={handleDisconnect}
              disabled={isDeleting}
            >
              {isDeleting
                ? t("googleWorkspaceConnector.actions.disconnecting")
                : t("googleWorkspaceConnector.actions.disconnect")}
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
                ? t("googleWorkspaceConnector.errors.bridgeDisabled")
                : t("googleWorkspaceConnector.errors.bridgeFailed")}
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
