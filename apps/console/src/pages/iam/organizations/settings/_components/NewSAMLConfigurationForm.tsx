// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useToast } from "@probo/ui";
import { useCallback } from "react";
import { ConnectionHandler, graphql } from "react-relay";

import type { NewSAMLConfigurationForm_createMutation } from "#/__generated__/iam/NewSAMLConfigurationForm_createMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  SAMLConfigurationForm,
  type SAMLConfigurationFormData,
} from "./SAMLConfigurationForm";

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
          canUpdate: permission(action: "iam:saml-configuration:update")
          canDelete: permission(action: "iam:saml-configuration:delete")
        }
      }
    }
  }
`;

export function NewSAMLConfigurationForm(props: { onCreate: () => void }) {
  const { onCreate } = props;
  const organizationId = useOrganizationId();

  const { __ } = useTranslate();
  const { toast } = useToast();

  const [create, isCreating]
    = useMutationWithToasts<NewSAMLConfigurationForm_createMutation>(
      createSAMLConfigurationMutation,
      {
        successMessage: "SAML configuration created successfully.",
        errorMessage: "Failed to create SAML configuration",
      },
    );

  const handleCreate = useCallback(
    async (data: SAMLConfigurationFormData) => {
      const connectionID = ConnectionHandler.getConnectionID(
        organizationId,
        "SAMLConfigurationListFragment_samlConfigurations",
      );

      await create({
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
        onCompleted: (_, e) => {
          if (e) {
            toast({
              variant: "error",
              title: __("Error"),
              description: formatError(
                __("Failed to create SAML configuration"),
                e,
              ),
            });
            return;
          }

          onCreate();
        },
      });
    },
    [organizationId, create, onCreate, __, toast],
  );

  return (
    <SAMLConfigurationForm onSubmit={handleCreate} disabled={isCreating} />
  );
}
