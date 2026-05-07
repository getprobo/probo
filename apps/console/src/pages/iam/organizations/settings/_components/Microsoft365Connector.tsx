// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  IconSettingsGear2,
  Input,
  Microsoft,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useState } from "react";
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
  const { __, dateTimeFormat } = useTranslate();
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
            title: __("Error"),
            description: errors.map(e => e.message).join(", "),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Microsoft 365 disconnected successfully"),
          variant: "success",
        });
        dialogRef.current?.close();
      },
      onError(error) {
        toast({
          title: __("Error"),
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
            title: __("Error"),
            description: errors.map(e => e.message).join(", "),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Excluded user names updated successfully"),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: __("Error"),
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
          <h3 className="font-medium">{__("Microsoft 365")}</h3>
          <p className="text-sm text-txt-secondary">
            {__(
              "Connect Microsoft 365 to automatically sync users via SCIM.",
            )}
          </p>
        </div>
        <Button variant="secondary" onClick={handleConnect}>
          {__("Connect")}
        </Button>
      </Card>
    );
  }

  return (
    <Card padded className="flex items-center gap-3">
      <div className="w-10 h-10 flex items-center justify-center bg-subtle rounded">
        <Microsoft className="w-6 h-6" />
      </div>
      <div className="mr-auto">
        <h3 className="font-medium">{__("Microsoft 365")}</h3>
        <p className="text-sm text-txt-secondary">
          {sprintf(__("Connected on %s"), dateTimeFormat(connector.createdAt))}
        </p>
      </div>
      <Badge variant="success" size="md">
        {__("Connected")}
      </Badge>
      <Dialog
        ref={excludedUserNamesDialogRef}
        trigger={(
          <Button variant="secondary">
            <IconSettingsGear2 size={16} />
            {__("Settings")}
          </Button>
        )}
        title={__("Microsoft 365 Settings")}
        className="max-w-lg"
      >
        <DialogContent padded className="space-y-6">
          <div className="space-y-4">
            <div>
              <h4 className="text-sm font-medium">{__("Excluded user names")}</h4>
              <p className="text-sm text-txt-secondary mt-1">
                {__("Users with these user names will not be synced from Microsoft 365.")}
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
                {__("Add")}
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
                      {__("Remove")}
                    </Button>
                  </div>
                ))}
              </div>
            )}

            {currentExcludedUserNames.length === 0 && (
              <p className="text-sm text-txt-secondary text-center py-4">
                {__("No excluded user names. All Microsoft 365 users will be synced.")}
              </p>
            )}
          </div>
        </DialogContent>
      </Dialog>
      <Dialog
        ref={dialogRef}
        trigger={(
          <Button variant="danger">
            {__("Disconnect")}
          </Button>
        )}
        title={__("Disconnect Microsoft 365")}
        className="max-w-lg"
      >
        <DialogContent padded className="space-y-4">
          <p className="text-txt-secondary text-sm">
            {__(
              "This will disconnect your Microsoft 365 integration. Users will no longer be automatically synced via SCIM.",
            )}
          </p>
          <p className="text-red-600 text-sm font-medium">
            {__("This action cannot be undone.")}
          </p>
        </DialogContent>
        <DialogFooter>
          <Button
            variant="danger"
            onClick={handleDisconnect}
            disabled={isDeleting}
          >
            {isDeleting
              ? __("Disconnecting...")
              : __("Disconnect")}
          </Button>
        </DialogFooter>
      </Dialog>
    </Card>
  );
}
