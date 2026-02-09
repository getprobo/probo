import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  Google,
  IconSettingsGear2,
  Input,
  useDialogRef,
} from "@probo/ui";
import { useState } from "react";
import { graphql, useFragment } from "react-relay";

import type { GoogleWorkspaceConnectorDeleteMutation } from "#/__generated__/iam/GoogleWorkspaceConnectorDeleteMutation.graphql";
import type { GoogleWorkspaceConnectorFragment$key } from "#/__generated__/iam/GoogleWorkspaceConnectorFragment.graphql";
import type { GoogleWorkspaceConnectorUpdateSCIMBridgeMutation } from "#/__generated__/iam/GoogleWorkspaceConnectorUpdateSCIMBridgeMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const googleWorkspaceConnectorFragment = graphql`
  fragment GoogleWorkspaceConnectorFragment on SCIMConfiguration {
    id
    bridge {
      id
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
      deletedScimConfigurationId
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
}) {
  const { fKey } = props;
  const data = useFragment<GoogleWorkspaceConnectorFragment$key>(googleWorkspaceConnectorFragment, fKey);
  const bridge = data?.bridge;
  const connector = bridge?.connector;
  const scimConfigurationId = data?.id;
  const bridgeId = bridge?.id;

  const organizationId = useOrganizationId();
  const { __, dateTimeFormat } = useTranslate();
  const dialogRef = useDialogRef();
  const excludedUserNamesDialogRef = useDialogRef();

  const [newUser, setNewUser] = useState("");

  const [deleteSCIMConfiguration, isDeleting]
    = useMutationWithToasts<GoogleWorkspaceConnectorDeleteMutation>(
      deleteSCIMConfigurationMutation,
      {
        successMessage: __("Google Workspace disconnected successfully"),
        errorMessage: __("Failed to disconnect Google Workspace"),
      },
    );

  const [updateSCIMBridge, isUpdating]
    = useMutationWithToasts<GoogleWorkspaceConnectorUpdateSCIMBridgeMutation>(
      updateSCIMBridgeMutation,
      {
        successMessage: __("Excluded user names updated successfully"),
        errorMessage: __("Failed to update excluded user names"),
      },
    );

  const handleConnect = () => {
    const baseUrl = import.meta.env.VITE_API_URL || window.location.origin;
    const url = new URL("/api/console/v1/connectors/initiate", baseUrl);
    url.searchParams.append("organization_id", organizationId);
    url.searchParams.append("provider", "GOOGLE_WORKSPACE");
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
      onCompleted: () => {
        dialogRef.current?.close();
      },
      updater: (store) => {
        const organizationRecord = store.get(organizationId);
        if (organizationRecord) {
          organizationRecord.setValue(null, "scimConfiguration");
        }
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
          <h3 className="font-medium">{__("Google Workspace")}</h3>
          <p className="text-sm text-txt-secondary">
            {__(
              "Connect Google Workspace to automatically sync users via SCIM.",
            )}
          </p>
        </div>
        <Button variant="secondary" onClick={handleConnect}>
          {__("Connect")}
        </Button>
      </Card>
    );
  }

  // Connected state
  return (
    <Card padded className="flex items-center gap-3">
      <div className="w-10 h-10 flex items-center justify-center bg-subtle rounded">
        <Google className="w-6 h-6" />
      </div>
      <div className="mr-auto">
        <h3 className="font-medium">{__("Google Workspace")}</h3>
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
        title={__("Google Workspace Settings")}
        className="max-w-lg"
      >
        <DialogContent padded className="space-y-6">
          <div className="space-y-4">
            <div>
              <h4 className="text-sm font-medium">{__("Excluded user names")}</h4>
              <p className="text-sm text-txt-secondary mt-1">
                {__("Users with these user names will not be synced from Google Workspace.")}
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
                {__("No excluded user names. All Google Workspace users will be synced.")}
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
        title={__("Disconnect Google Workspace")}
        className="max-w-lg"
      >
        <DialogContent padded className="space-y-4">
          <p className="text-txt-secondary text-sm">
            {__(
              "This will disconnect your Google Workspace integration. Users will no longer be automatically synced via SCIM.",
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
