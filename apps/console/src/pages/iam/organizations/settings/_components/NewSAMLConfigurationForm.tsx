// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
