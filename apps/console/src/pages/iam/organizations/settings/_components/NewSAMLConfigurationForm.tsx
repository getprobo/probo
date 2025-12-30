import { ConnectionHandler, graphql } from "react-relay";
import type { NewSAMLConfigurationForm_createMutation } from "/__generated__/iam/NewSAMLConfigurationForm_createMutation.graphql";
import {
  SAMLConfigurationForm,
  type SAMLConfigurationFormData,
} from "./SAMLConfigurationForm";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useCallback } from "react";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";

const createSAMLConfigurationMutation = graphql`
  mutation NewSAMLConfigurationForm_createMutation(
    $input: CreateSAMLConfigurationInput!
    $connections: [ID!]!
  ) {
    createSAMLConfiguration(input: $input) {
      samlConfigurationEdge @prependEdge(connections: $connections) {
        node {
          id
          emailDomain
          enforcementPolicy
          domainVerificationToken
          domainVerifiedAt
          testLoginUrl
        }
      }
    }
  }
`;

export function NewSAMLConfigurationForm(props: { onCreate: () => void }) {
  const { onCreate } = props;
  const organizationId = useOrganizationId();

  const [create, isCreating] =
    useMutationWithToasts<NewSAMLConfigurationForm_createMutation>(
      createSAMLConfigurationMutation,
      {
        successMessage: "SAML configuration created successfully.",
        errorMessage: "Failed to create SAML configuration",
      },
    );

  const handleCreate = useCallback(
    (data: SAMLConfigurationFormData) => {
      const connectionID = ConnectionHandler.getConnectionID(
        organizationId,
        "SAMLConfigurationListFragment_samlConfigurations",
      );

      create({
        variables: {
          input: {
            organizationId,
            emailDomain: data.emailDomain,
            idpEntityId: data.idpEntityId,
            idpSsoUrl: data.idpSsoUrl,
            idpCertificate: data.idpCertificate,
            autoSignupEnabled: data.autoSignupEnabled,
            attributeMappings: data.attributeMappings,
          },
          connections: [connectionID],
        },
        onCompleted: onCreate,
      });
    },
    [organizationId, create, onCreate],
  );

  return (
    <SAMLConfigurationForm onSubmit={handleCreate} disabled={isCreating} />
  );
}
