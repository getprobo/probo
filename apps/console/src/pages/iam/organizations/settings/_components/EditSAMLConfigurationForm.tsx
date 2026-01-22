import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useToast } from "@probo/ui";
import { useCallback } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { EditSAMLConfigurationForm_updateMutation } from "#/__generated__/iam/EditSAMLConfigurationForm_updateMutation.graphql";
import type { EditSAMLConfigurationFormQuery } from "#/__generated__/iam/EditSAMLConfigurationFormQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  SAMLConfigurationForm,
  type SAMLConfigurationFormData,
} from "./SAMLConfigurationForm";

export const samlConfigurationFormQuery = graphql`
  query EditSAMLConfigurationFormQuery($samlConfigurationId: ID!) {
    samlConfiguration: node(id: $samlConfigurationId) @required(action: THROW) {
      __typename
      ... on SAMLConfiguration {
        id
        emailDomain
        enforcementPolicy
        domainVerificationToken
        domainVerifiedAt
        testLoginUrl
        idpEntityId
        idpSsoUrl
        idpCertificate
        attributeMappings {
          email
          firstName
          lastName
          role
        }
        autoSignupEnabled
      }
    }
  }
`;

const updateSAMLConfigurationMutation = graphql`
  mutation EditSAMLConfigurationForm_updateMutation(
    $input: UpdateSAMLConfigurationInput!
  ) {
    updateSAMLConfiguration(input: $input) {
      samlConfiguration {
        id
        emailDomain
        enforcementPolicy
        domainVerificationToken
        domainVerifiedAt
        testLoginUrl
      }
    }
  }
`;

export function EditSAMLConfigurationForm(props: {
  onUpdate: () => void;
  queryRef: PreloadedQuery<EditSAMLConfigurationFormQuery>;
}) {
  const { onUpdate, queryRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const { toast } = useToast();

  const { samlConfiguration }
    = usePreloadedQuery<EditSAMLConfigurationFormQuery>(
      samlConfigurationFormQuery,
      queryRef,
    );
  if (samlConfiguration.__typename !== "SAMLConfiguration") {
    throw new Error("node is not a SAML configuration");
  }

  const [update, isUpdating]
    = useMutationWithToasts<EditSAMLConfigurationForm_updateMutation>(
      updateSAMLConfigurationMutation,
      {
        successMessage: "SAML configuration updated successfully.",
        errorMessage: "Failed to update SAML configuration. Please try again.",
      },
    );

  const handleUpdate = useCallback(
    async (data: SAMLConfigurationFormData) => {
      await update({
        variables: {
          input: {
            samlConfigurationId: samlConfiguration.id,
            organizationId,
            idpEntityId: data.idpEntityId,
            idpSsoUrl: data.idpSsoUrl,
            idpCertificate: data.idpCertificate,
            autoSignupEnabled: data.autoSignupEnabled,
            enforcementPolicy: data.enforcementPolicy,
            attributeMappings: data.attributeMappings,
          },
        },
        onCompleted: (_, e) => {
          if (e) {
            toast({
              variant: "error",
              title: __("Error"),
              description: formatError(
                __("Failed to update SAML configuration"),
                e,
              ),
            });
            return;
          }

          onUpdate();
        },
      });
    },
    [onUpdate, organizationId, samlConfiguration.id, update, __, toast],
  );

  return (
    <SAMLConfigurationForm
      disabled={isUpdating}
      initialValues={samlConfiguration}
      isEditing
      onSubmit={handleUpdate}
    />
  );
}
