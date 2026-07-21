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

import { formatError } from "@probo/helpers";
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
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();
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
            title: t("scimConfiguration.errors.title"),
            description: formatError(
              t("scimConfiguration.errors.create"),
              e,
            ),
          });
          return;
        }

        if (response.createSCIMConfiguration) {
          setToken(response.createSCIMConfiguration.token);
        }
        toast({
          title: t("scimConfiguration.messages.configured.title"),
          description: t("scimConfiguration.messages.copyToken"),
          variant: "success",
        });
      },
      onError: (error: Error) => {
        toast({
          variant: "error",
          title: t("scimConfiguration.errors.title"),
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
          title: t("scimConfiguration.messages.deleted.title"),
          description: t("scimConfiguration.messages.deleted.description"),
          variant: "success",
        });
      },
      onError: (error: Error) => {
        toast({
          variant: "error",
          title: t("scimConfiguration.errors.title"),
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
          title: t("scimConfiguration.messages.regenerated.title"),
          description: t("scimConfiguration.messages.regenerated.description"),
          variant: "success",
        });
      },
      onError: (error: Error) => {
        toast({
          variant: "error",
          title: t("scimConfiguration.errors.title"),
          description: error.message,
        });
      },
    });
  };

  const copyToClipboard = (text: string, label: string) => {
    void navigator.clipboard.writeText(text);
    toast({
      title: t("scimConfiguration.messages.copied"),
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
            <h3 className="font-medium">
              {t("scimConfiguration.empty.title")}
            </h3>
            <p className="text-sm text-txt-secondary mt-1">
              {t("scimConfiguration.empty.description")}
            </p>
          </div>
          {canCreate && (
            <Button
              onClick={handleCreate}
              disabled={isCreatingSAMLConfiguration}
            >
              {isCreatingSAMLConfiguration
                ? t("scimConfiguration.actions.enabling")
                : t("scimConfiguration.actions.enable")}
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
              <h3 className="font-medium">
                {t("scimConfiguration.active.title")}
              </h3>
              <p className="text-sm text-txt-secondary">
                {t("scimConfiguration.active.description")}
              </p>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">
                {t("scimConfiguration.fields.endpointUrl")}
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
                      t("scimConfiguration.fields.endpointUrl"),
                    )}
                  icon={IconSquareBehindSquare2}
                />
              </div>
            </div>

            {token && (
              <div>
                <label className="text-sm font-medium">
                  {t("scimConfiguration.fields.bearerToken")}
                </label>
                <p className="text-xs text-txt-warning mb-1">
                  {t("scimConfiguration.tokenWarning")}
                </p>
                <div className="flex items-center gap-2 mt-1">
                  <code className="flex-1 bg-subtle p-2 rounded text-sm font-mono break-all">
                    {token}
                  </code>
                  <Button
                    variant="secondary"
                    onClick={() =>
                      copyToClipboard(
                        token,
                        t("scimConfiguration.fields.bearerToken"),
                      )}
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
                  ? t("scimConfiguration.actions.regenerating")
                  : t("scimConfiguration.actions.regenerateToken")}
              </Button>
              {canDelete && (
                <Button
                  variant="danger"
                  onClick={() => deleteDialogRef.current?.open()}
                  icon={IconTrashCan}
                >
                  {t("scimConfiguration.actions.deleteConfiguration")}
                </Button>
              )}
            </div>
          </div>
        </div>
      </Card>

      <Dialog
        ref={deleteDialogRef}
        title={t("scimConfiguration.delete.title")}
        onClose={() => deleteDialogRef.current?.close()}
      >
        <div className="p-4 space-y-4">
          <p>
            {t("scimConfiguration.delete.description")}
          </p>
          <ul className="list-disc list-inside text-sm space-y-1">
            <li>{t("scimConfiguration.delete.effects.disable")}</li>
            <li>
              {t("scimConfiguration.delete.effects.changeMembershipSource")}
            </li>
            <li>{t("scimConfiguration.delete.effects.invalidateToken")}</li>
          </ul>
          <p className="text-sm text-txt-secondary">
            {t("scimConfiguration.delete.note")}
          </p>
          <div className="flex justify-end gap-2">
            <Button
              variant="secondary"
              onClick={() => deleteDialogRef.current?.close()}
            >
              {t("scimConfiguration.actions.cancel")}
            </Button>
            <Button
              variant="danger"
              onClick={handleDelete}
              disabled={isDeletingSCIMConfiguration}
            >
              {isDeletingSCIMConfiguration
                ? t("scimConfiguration.actions.deleting")
                : t("scimConfiguration.actions.delete")}
            </Button>
          </div>
        </div>
      </Dialog>
    </>
  );
}
