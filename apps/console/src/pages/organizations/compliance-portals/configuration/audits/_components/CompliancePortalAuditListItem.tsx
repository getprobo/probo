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

import { getAuditStateVariant, getCompliancePortalLinkedVisibilityOptions } from "@probo/helpers";
import { dateFormat } from "@probo/i18n";
import { Badge, Button, Field, IconCrossLargeX, Option, Td, Tr } from "@probo/ui";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type {
  CompliancePortalAuditListItem_catalogAuditFragment$key,
} from "#/__generated__/core/CompliancePortalAuditListItem_catalogAuditFragment.graphql";
import type {
  CompliancePortalAuditListItem_compliancePortalFragment$key,
} from "#/__generated__/core/CompliancePortalAuditListItem_compliancePortalFragment.graphql";
import type {
  CompliancePortalAuditListItem_removeMutation,
} from "#/__generated__/core/CompliancePortalAuditListItem_removeMutation.graphql";
import type {
  CompliancePortalAuditListItem_updateAuditVisibilityMutation,
} from "#/__generated__/core/CompliancePortalAuditListItem_updateAuditVisibilityMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const compliancePortalFragment = graphql`
  fragment CompliancePortalAuditListItem_compliancePortalFragment on CompliancePortal {
    id
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

const catalogAuditFragment = graphql`
  fragment CompliancePortalAuditListItem_catalogAuditFragment on CompliancePortalAudit {
    id
    visibility
    audit {
      id
      name
      framework {
        name
      }
      validUntil
      state
    }
  }
`;

const updateAuditVisibilityMutation = graphql`
  mutation CompliancePortalAuditListItem_updateAuditVisibilityMutation(
    $input: UpdateCompliancePortalAuditVisibilityInput!
  ) {
    updateCompliancePortalAuditVisibility(input: $input) {
      catalogAudit {
        id
        visibility
      }
    }
  }
`;

const removeAuditMutation = graphql`
  mutation CompliancePortalAuditListItem_removeMutation(
    $input: DeleteCompliancePortalAuditInput!
    $connections: [ID!]!
  ) {
    deleteCompliancePortalAudit(input: $input) {
      deletedCompliancePortalAuditId @deleteEdge(connections: $connections)
    }
  }
`;

export function CompliancePortalAuditListItem(props: {
  catalogAuditFragmentRef: CompliancePortalAuditListItem_catalogAuditFragment$key;
  compliancePortalFragmentRef: CompliancePortalAuditListItem_compliancePortalFragment$key;
  auditsConnectionId: DataID;
}) {
  const { catalogAuditFragmentRef, compliancePortalFragmentRef, auditsConnectionId } = props;

  const organizationId = useOrganizationId();
  const { i18n, t } = useTranslation("organizations/compliance-portals");

  const compliancePortal = useFragment<CompliancePortalAuditListItem_compliancePortalFragment$key>(
    compliancePortalFragment,
    compliancePortalFragmentRef,
  );
  const catalogAudit = useFragment<CompliancePortalAuditListItem_catalogAuditFragment$key>(
    catalogAuditFragment,
    catalogAuditFragmentRef,
  );
  const audit = catalogAudit.audit;

  const [updateAuditVisibility, isUpdatingAuditVisibility] = useMutation<
    CompliancePortalAuditListItem_updateAuditVisibilityMutation
  >(
    updateAuditVisibilityMutation,
    {
      successMessage: t("auditListItem.messages.visibilityUpdated"),
      errorToast: t("auditListItem.errors.updateVisibility"),
    },
  );
  const [removeAudit, isRemoving] = useMutation<CompliancePortalAuditListItem_removeMutation>(
    removeAuditMutation,
    {
      successMessage: t("auditListItem.messages.removed"),
      errorToast: t("auditListItem.errors.remove"),
    },
  );

  const handleVisibilityChange = useCallback(
    async (value: string) => {
      const typedValue = value === "PUBLIC" ? "PUBLIC" : "RESTRICTED";
      await updateAuditVisibility({
        variables: {
          input: {
            compliancePortalId: compliancePortal.id,
            auditId: audit.id,
            compliancePortalVisibility: typedValue,
          },
        },
      });
    },
    [compliancePortal.id, audit.id, updateAuditVisibility],
  );

  const handleRemove = useCallback(async () => {
    await removeAudit({
      variables: {
        connections: [auditsConnectionId],
        input: {
          id: catalogAudit.id,
        },
      },
    });
  }, [auditsConnectionId, catalogAudit.id, removeAudit]);

  const visibilityOptions = getCompliancePortalLinkedVisibilityOptions(t);
  const validUntilFormatted = audit.validUntil
    ? dateFormat(i18n.language, audit.validUntil)
    : t("auditListItem.noExpiry");

  return (
    <Tr to={`/organizations/${organizationId}/audits/${audit.id}`}>
      <Td>
        <div className="flex gap-4 items-center">{audit.framework?.name}</div>
      </Td>
      <Td>{audit.name || t("auditListItem.untitled")}</Td>
      <Td>{validUntilFormatted}</Td>
      <Td>
        <Badge variant={getAuditStateVariant(audit.state)}>
          {t(`auditListItem.states.${audit.state.toLowerCase()}`)}
        </Badge>
      </Td>
      <Td noLink width={130} className="pr-0">
        <Field
          type="select"
          value={catalogAudit.visibility}
          onValueChange={value => void handleVisibilityChange(value)}
          disabled={isUpdatingAuditVisibility || !compliancePortal.canUpdate}
          className="w-[105px]"
        >
          {visibilityOptions.map(option => (
            <Option key={option.value} value={option.value}>
              <div className="flex items-center justify-between w-full">
                <Badge variant={option.variant}>{option.label}</Badge>
              </div>
            </Option>
          ))}
        </Field>
      </Td>
      <Td noLink width={48}>
        {compliancePortal.canUpdate && (
          <Button
            variant="tertiary"
            icon={IconCrossLargeX}
            aria-label={t("auditListItem.actions.remove")}
            disabled={isRemoving}
            onClick={() => void handleRemove()}
          />
        )}
      </Td>
    </Tr>
  );
}
