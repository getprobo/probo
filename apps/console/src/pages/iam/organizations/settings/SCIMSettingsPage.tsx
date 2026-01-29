import { useTranslate } from "@probo/i18n";
import { Spinner } from "@probo/ui";
import { useEffect, useRef } from "react";
import { graphql, type PreloadedQuery,
  useMutation, usePreloadedQuery } from "react-relay";
import { useSearchParams } from "react-router";

import type { SCIMSettingsPageCreateSCIMConfigurationMutation } from "#/__generated__/iam/SCIMSettingsPageCreateSCIMConfigurationMutation.graphql";
import type { SCIMSettingsPageQuery } from "#/__generated__/iam/SCIMSettingsPageQuery.graphql";

import { ConnectorList } from "./_components/ConnectorList";
import { SCIMConfiguration } from "./_components/SCIMConfiguration";
import { SCIMEventList } from "./_components/SCIMEventList";

export const scimSettingsPageQuery = graphql`
  query SCIMSettingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        id

        scimConfiguration {
          id
          bridge {
            id
          }
          ...SCIMEventListFragment
        }

        ...SCIMConfigurationFragment
        ...ConnectorListFragment
      }
    }
  }
`;

const createSCIMConfigurationMutation = graphql`
  mutation SCIMSettingsPageCreateSCIMConfigurationMutation(
    $input: CreateSCIMConfigurationInput!
  ) {
    createSCIMConfiguration(input: $input) {
      scimConfiguration {
        id
      }
      scimBridge {
        id
      }
    }
  }
`;

export function SCIMSettingsPage(props: {
  queryRef: PreloadedQuery<SCIMSettingsPageQuery>;
}) {
  const { queryRef } = props;
  const { __ } = useTranslate();
  const [searchParams, setSearchParams] = useSearchParams();
  const connectorId = searchParams.get("connector_id");
  const mutationTriggeredRef = useRef(false);

  const { organization } = usePreloadedQuery(scimSettingsPageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid node type");
  }

  const [createSCIMConfiguration]
    = useMutation<SCIMSettingsPageCreateSCIMConfigurationMutation>(
      createSCIMConfigurationMutation,
    );

  // Auto-create SCIM configuration and bridge when connector_id is in URL
  useEffect(() => {
    if (!connectorId || mutationTriggeredRef.current) return;

    // Don't create if SCIM config already exists
    if (organization.scimConfiguration?.id) {
      setSearchParams((params: URLSearchParams) => {
        params.delete("connector_id");
        return params;
      });
      return;
    }

    mutationTriggeredRef.current = true;

    createSCIMConfiguration({
      variables: {
        input: {
          organizationId: organization.id,
          connectorId: connectorId,
        },
      },
      onCompleted: () => {
        const url = new URL(window.location.href);
        url.searchParams.delete("connector_id");
        window.location.href = url.toString();
      },
      onError: (error) => {
        console.error("Failed to create SCIM configuration:", error);
        mutationTriggeredRef.current = false;
        setSearchParams((params: URLSearchParams) => {
          params.delete("connector_id");
          return params;
        });
      },
    });
  }, [
    connectorId,
    organization.id,
    organization.scimConfiguration?.id,
    createSCIMConfiguration,
    setSearchParams,
  ]);

  // Show loader while creating SCIM configuration
  if (connectorId) {
    return (
      <div className="flex items-center justify-center py-12">
        <Spinner size={24} />
      </div>
    );
  }

  // Check if connected via Identity Provider (SCIM config has a bridge)
  const hasIdentityProvider = !!organization.scimConfiguration?.bridge;

  // Check if Manual SCIM is configured (SCIM config exists but no bridge)
  const hasManualScim = !!organization.scimConfiguration && !organization.scimConfiguration.bridge;

  // Show Identity Provider section when:
  // - No SCIM config yet (user can connect)
  // - Or SCIM config with bridge (already connected via IdP)
  const showIdentityProviderSection = !organization.scimConfiguration || hasIdentityProvider;

  // Show Manual SCIM section when:
  // - Manual SCIM is configured (no bridge)
  // - Or no SCIM config at all (user can choose to enable manual)
  const showManualScimSection = hasManualScim || !organization.scimConfiguration;

  // Show provisioning events when SCIM is configured (either manual or via IdP)
  const showProvisioningEvents = !!organization.scimConfiguration;

  return (
    <div className="space-y-8">
      {showIdentityProviderSection && (
        <ConnectorList fKey={organization} />
      )}

      {showManualScimSection && (
        <div className="space-y-4">
          <h2 className="text-base font-medium">{__("Manual SCIM")}</h2>
          {!hasManualScim && (
            <p className="text-sm text-txt-secondary">
              {__(
                "Configure SCIM manually if your identity provider is not listed above. This requires setting up the SCIM endpoint URL and bearer token in your identity provider.",
              )}
            </p>
          )}
          <SCIMConfiguration fKey={organization} />
        </div>
      )}

      {showProvisioningEvents && (
        <div className="space-y-4">
          <h2 className="text-base font-medium">
            {__("Provisioning Event History")}
          </h2>
          <SCIMEventList fKey={organization.scimConfiguration} />
        </div>
      )}
    </div>
  );
}
