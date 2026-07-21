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
import { useToast } from "@probo/ui";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
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
        # eslint-disable-next-line relay/unused-fields
        emailDomain
        enforcementPolicy
        # eslint-disable-next-line relay/unused-fields
        domainVerificationToken
        # eslint-disable-next-line relay/unused-fields
        domainVerifiedAt
        # eslint-disable-next-line relay/unused-fields
        testLoginUrl
        idpEntityId
        idpSsoUrl
        idpCertificate
        attributeMappings {
          # eslint-disable-next-line relay/unused-fields
          email
          # eslint-disable-next-line relay/unused-fields
          firstName
          # eslint-disable-next-line relay/unused-fields
          lastName
          # eslint-disable-next-line relay/unused-fields
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
  const { t } = useTranslation();
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
        successMessage: t("editSamlConfigurationForm.messages.updated"),
        errorMessage: t("editSamlConfigurationForm.errors.update"),
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
              title: t("common.error"),
              description: formatError(
                t("editSamlConfigurationForm.errors.update"),
                e,
              ),
            });
            return;
          }

          onUpdate();
        },
      });
    },
    [onUpdate, organizationId, samlConfiguration.id, update, t, toast],
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
