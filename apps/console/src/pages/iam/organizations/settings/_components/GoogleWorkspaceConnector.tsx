import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  useDialogRef,
} from "@probo/ui";
import { graphql, useFragment } from "react-relay";

import type { GoogleWorkspaceConnectorDeleteMutation } from "#/__generated__/iam/GoogleWorkspaceConnectorDeleteMutation.graphql";
import type { GoogleWorkspaceConnectorFragment$key } from "#/__generated__/iam/GoogleWorkspaceConnectorFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const googleWorkspaceConnectorFragment = graphql`
  fragment GoogleWorkspaceConnectorFragment on SCIMConfiguration {
    id
    bridge {
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

export function GoogleWorkspaceConnector(props: {
  fKey: GoogleWorkspaceConnectorFragment$key | null;
}) {
  const { fKey } = props;
  const data = useFragment<GoogleWorkspaceConnectorFragment$key>(googleWorkspaceConnectorFragment, fKey);
  const connector = data?.bridge?.connector;
  const scimConfigurationId = data?.id;

  const organizationId = useOrganizationId();
  const { __, dateTimeFormat } = useTranslate();
  const dialogRef = useDialogRef();

  const [deleteSCIMConfiguration, isDeleting]
    = useMutationWithToasts<GoogleWorkspaceConnectorDeleteMutation>(
      deleteSCIMConfigurationMutation,
      {
        successMessage: __("Google Workspace disconnected successfully"),
        errorMessage: __("Failed to disconnect Google Workspace"),
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

  // Not connected state
  if (!connector) {
    return (
      <Card padded className="flex items-center gap-3">
        <div className="w-10 h-10 flex items-center justify-center bg-subtle rounded">
          <img
            src="/google-workspace.png"
            alt="Google Workspace"
            className="w-6 h-6"
          />
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
        <img
          src="/google-workspace.png"
          alt="Google Workspace"
          className="w-6 h-6"
        />
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
        ref={dialogRef}
        trigger={(
          <Button variant="secondary">
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
            {isDeleting ? __("Disconnecting...") : __("Disconnect")}
          </Button>
        </DialogFooter>
      </Dialog>
    </Card>
  );
}
