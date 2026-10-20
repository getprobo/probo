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

import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { EditSAMLConfigurationForm_samlConfiguration$key } from "#/__generated__/iam/EditSAMLConfigurationForm_samlConfiguration.graphql";
import type { EditSAMLConfigurationForm_updateMutation } from "#/__generated__/iam/EditSAMLConfigurationForm_updateMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  SAMLConfigurationForm,
  type SAMLConfigurationFormData,
} from "./SAMLConfigurationForm";

const editSAMLConfigurationFormFragment = graphql`
  fragment EditSAMLConfigurationForm_samlConfiguration on SAMLConfiguration {
    id
    emailDomain
    enforcementPolicy
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
    canUpdate: permission(action: "iam:saml-configuration:update")
  }
`;

const updateSAMLConfigurationMutation = graphql`
  mutation EditSAMLConfigurationForm_updateMutation(
    $input: UpdateSAMLConfigurationInput!
  ) {
    updateSAMLConfiguration(input: $input) {
      samlConfiguration {
        ...EditSAMLConfigurationForm_samlConfiguration
        ...SAMLConfigurationListItem_samlConfiguration
      }
    }
  }
`;

export function EditSAMLConfigurationForm({
  samlConfigurationKey,
  ssoLoginUrl,
}: {
  samlConfigurationKey: EditSAMLConfigurationForm_samlConfiguration$key;
  ssoLoginUrl?: string | null;
}) {
  const samlConfiguration = useFragment(
    editSAMLConfigurationFormFragment,
    samlConfigurationKey,
  );
  const { canUpdate } = samlConfiguration;

  const organizationId = useOrganizationId();
  const { t } = useTranslation();

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
      });
    },
    [organizationId, samlConfiguration.id, update],
  );

  return (
    <SAMLConfigurationForm
      disabled={!canUpdate || isUpdating}
      hideSubmit={!canUpdate}
      initialValues={{
        emailDomain: samlConfiguration.emailDomain,
        enforcementPolicy: samlConfiguration.enforcementPolicy,
        idpEntityId: samlConfiguration.idpEntityId,
        idpSsoUrl: samlConfiguration.idpSsoUrl,
        idpCertificate: samlConfiguration.idpCertificate,
        attributeMappings: {
          email: samlConfiguration.attributeMappings.email,
          firstName: samlConfiguration.attributeMappings.firstName,
          lastName: samlConfiguration.attributeMappings.lastName,
          role: samlConfiguration.attributeMappings.role,
        },
        autoSignupEnabled: samlConfiguration.autoSignupEnabled,
      }}
      isEditing
      ssoLoginUrl={ssoLoginUrl}
      onSubmit={handleUpdate}
    />
  );
}
