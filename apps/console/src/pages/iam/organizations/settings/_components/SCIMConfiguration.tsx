import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  Dialog,
  IconRotateCw,
  IconSquareBehindSquare2,
  IconTrashCan,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { graphql, useFragment, useMutation } from "react-relay";

import type { SCIMConfigurationCreateMutation } from "#/__generated__/iam/SCIMConfigurationCreateMutation.graphql";
import type { SCIMConfigurationDeleteMutation } from "#/__generated__/iam/SCIMConfigurationDeleteMutation.graphql";
import type { SCIMConfigurationFragment$key } from "#/__generated__/iam/SCIMConfigurationFragment.graphql";
import type { SCIMConfigurationRegenerateTokenMutation } from "#/__generated__/iam/SCIMConfigurationRegenerateTokenMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const SCIMConfigurationFragment = graphql`
  fragment SCIMConfigurationFragment on Organization {
    canCreateSCIMConfiguration: permission(
      action: "iam:scim-configuration:create"
    )
    canDeleteSCIMConfiguration: permission(
      action: "iam:scim-configuration:delete"
    )
    scimConfiguration {
      id
      endpointUrl
      bridge {
        id
      }
    }
  }
`;

const createSCIMConfigurationMutation = graphql`
  mutation SCIMConfigurationCreateMutation(
    $input: CreateSCIMConfigurationInput!
  ) {
    createSCIMConfiguration(input: $input) {
      scimConfiguration {
        id
        endpointUrl

        organization {
          id
          scimConfiguration {
            id
            endpointUrl
          }
        }
      }
      token
    }
  }
`;

const deleteSCIMConfigurationMutation = graphql`
  mutation SCIMConfigurationDeleteMutation(
    $input: DeleteSCIMConfigurationInput!
  ) {
    deleteSCIMConfiguration(input: $input) {
      deletedScimConfigurationId @deleteRecord
    }
  }
`;

const regenerateSCIMTokenMutation = graphql`
  mutation SCIMConfigurationRegenerateTokenMutation(
    $input: RegenerateSCIMTokenInput!
  ) {
    regenerateSCIMToken(input: $input) {
      scimConfiguration {
        id
        endpointUrl
        createdAt
        updatedAt
      }
      token
    }
  }
`;

export function SCIMConfiguration(props: {
  fKey: SCIMConfigurationFragment$key;
}) {
  const { fKey } = props;

  const organizationId = useOrganizationId();

  const organization = useFragment<SCIMConfigurationFragment$key>(SCIMConfigurationFragment, fKey);
  const {
    canCreateSCIMConfiguration: canCreate,
    canDeleteSCIMConfiguration: canDelete,
    scimConfiguration,
  } = organization;
  const hasIdentityProvider = !!scimConfiguration?.bridge;
  const { __ } = useTranslate();
  const { toast } = useToast();

  const [token, setToken] = useState<string | null>(null);

  const deleteDialogRef = useDialogRef();

  const [createSCIMConfiguration, isCreatingSAMLConfiguration]
    = useMutation<SCIMConfigurationCreateMutation>(
      createSCIMConfigurationMutation,
    );
  const [deleteSCIMConfiguration, isDeletingSCIMConfiguration]
    = useMutation<SCIMConfigurationDeleteMutation>(
      deleteSCIMConfigurationMutation,
    );
  const [regenerateSCIMToken, isRegeneratingSCIMToken]
    = useMutation<SCIMConfigurationRegenerateTokenMutation>(
      regenerateSCIMTokenMutation,
    );

  const handleCreate = () => {
    createSCIMConfiguration({
      variables: {
        input: {
          organizationId,
        },
      },
      onCompleted: (response, e) => {
        if (e) {
          toast({
            variant: "error",
            title: __("Error"),
            description: formatError(
              __("Manual SCIM configuration failed"),
              e,
            ),
          });
          return;
        }

        if (response.createSCIMConfiguration) {
          setToken(response.createSCIMConfiguration.token);
        }
        toast({
          title: __("Manual SCIM Configured"),
          description: __(
            "Copy the bearer token now. It will not be shown again.",
          ),
          variant: "success",
        });
      },
      onError: (error: Error) => {
        toast({
          variant: "error",
          title: __("Error"),
          description: error.message,
        });
      },
    });
  };

  const handleDelete = () => {
    if (!scimConfiguration) return;

    deleteSCIMConfiguration({
      variables: {
        input: {
          organizationId,
          scimConfigurationId: scimConfiguration.id,
        },
      },
      onCompleted: () => {
        deleteDialogRef.current?.close();
        setToken(null);
        toast({
          title: __("Manual SCIM Configuration Deleted"),
          description: __(
            "All SCIM-provisioned memberships have been changed to manual source.",
          ),
          variant: "success",
        });
      },
      onError: (error: Error) => {
        toast({
          variant: "error",
          title: __("Error"),
          description: error.message,
        });
      },
    });
  };

  const handleRegenerate = () => {
    if (!scimConfiguration) return;

    regenerateSCIMToken({
      variables: {
        input: {
          organizationId,
          scimConfigurationId: scimConfiguration.id,
        },
      },
      onCompleted: (response) => {
        if (response.regenerateSCIMToken) {
          setToken(response.regenerateSCIMToken.token);
        }
        toast({
          title: __("Bearer Token Regenerated"),
          description: __(
            "Copy the new bearer token now. It will not be shown again.",
          ),
          variant: "success",
        });
      },
      onError: (error: Error) => {
        toast({
          variant: "error",
          title: __("Error"),
          description: error.message,
        });
      },
    });
  };

  const copyToClipboard = (text: string, label: string) => {
    void navigator.clipboard.writeText(text);
    toast({
      title: __("Copied to clipboard"),
      description: label,
      variant: "success",
    });
  };

  if (hasIdentityProvider) {
    return null;
  }

  if (!scimConfiguration) {
    return (
      <Card padded>
        <div className="flex items-center justify-between">
          <div>
            <h3 className="font-medium">{__("Manual SCIM is not configured")}</h3>
            <p className="text-sm text-txt-secondary mt-1">
              {__(
                "Generate a SCIM endpoint and bearer token to configure your identity provider manually.",
              )}
            </p>
          </div>
          {canCreate && (
            <Button
              onClick={handleCreate}
              disabled={isCreatingSAMLConfiguration}
            >
              {isCreatingSAMLConfiguration
                ? __("Enabling...")
                : __("Enable SCIM")}
            </Button>
          )}
        </div>
      </Card>
    );
  }

  return (
    <>
      <Card padded>
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="font-medium">{__("Manual SCIM Active")}</h3>
              <p className="text-sm text-txt-secondary">
                {__(
                  "Use these credentials to configure SCIM in your identity provider.",
                )}
              </p>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">
                {__("SCIM Endpoint URL")}
              </label>
              <div className="flex items-center gap-2 mt-1">
                <code className="flex-1 bg-subtle p-2 rounded text-sm font-mono">
                  {scimConfiguration.endpointUrl}
                </code>
                <Button
                  variant="secondary"
                  onClick={() =>
                    copyToClipboard(
                      scimConfiguration.endpointUrl,
                      __("SCIM Endpoint URL"),
                    )}
                  icon={IconSquareBehindSquare2}
                />
              </div>
            </div>

            {token && (
              <div>
                <label className="text-sm font-medium">
                  {__("Bearer Token")}
                </label>
                <p className="text-xs text-txt-warning mb-1">
                  {__("This token will only be shown once. Copy it now.")}
                </p>
                <div className="flex items-center gap-2 mt-1">
                  <code className="flex-1 bg-subtle p-2 rounded text-sm font-mono break-all">
                    {token}
                  </code>
                  <Button
                    variant="secondary"
                    onClick={() => copyToClipboard(token, __("Bearer Token"))}
                    icon={IconSquareBehindSquare2}
                  />
                </div>
              </div>
            )}

            <div className="flex items-center gap-2 pt-4 border-t border-border-low">
              <Button
                variant="secondary"
                onClick={handleRegenerate}
                disabled={isRegeneratingSCIMToken}
                icon={IconRotateCw}
              >
                {isRegeneratingSCIMToken
                  ? __("Regenerating...")
                  : __("Regenerate Token")}
              </Button>
              {canDelete && (
                <Button
                  variant="danger"
                  onClick={() => deleteDialogRef.current?.open()}
                  icon={IconTrashCan}
                >
                  {__("Delete Configuration")}
                </Button>
              )}
            </div>
          </div>
        </div>
      </Card>

      <Dialog
        ref={deleteDialogRef}
        title={__("Delete Manual SCIM Configuration")}
        onClose={() => deleteDialogRef.current?.close()}
      >
        <div className="p-4 space-y-4">
          <p>
            {__(
              "Are you sure you want to delete the manual SCIM configuration? This will:",
            )}
          </p>
          <ul className="list-disc list-inside text-sm space-y-1">
            <li>{__("Disable manual SCIM provisioning")}</li>
            <li>
              {__("Change all SCIM-provisioned memberships to manual source")}
            </li>
            <li>{__("Invalidate the current bearer token")}</li>
          </ul>
          <p className="text-sm text-txt-secondary">
            {__(
              "Existing users will not be removed, only their membership source will change.",
            )}
          </p>
          <div className="flex justify-end gap-2">
            <Button
              variant="secondary"
              onClick={() => deleteDialogRef.current?.close()}
            >
              {__("Cancel")}
            </Button>
            <Button
              variant="danger"
              onClick={handleDelete}
              disabled={isDeletingSCIMConfiguration}
            >
              {isDeletingSCIMConfiguration ? __("Deleting...") : __("Delete")}
            </Button>
          </div>
        </div>
      </Dialog>
    </>
  );
}
