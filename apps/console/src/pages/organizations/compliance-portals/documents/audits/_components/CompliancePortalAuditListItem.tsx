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
import { Badge, Checkbox, Field, Option, Td, Tr } from "@probo/ui";
import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type {
  CompliancePortalAuditListItem_audit$key,
} from "#/__generated__/core/CompliancePortalAuditListItem_audit.graphql";
import type {
  CompliancePortalAuditListItem_compliancePortal$key,
} from "#/__generated__/core/CompliancePortalAuditListItem_compliancePortal.graphql";
import type {
  CompliancePortalAuditListItem_removeMutation,
} from "#/__generated__/core/CompliancePortalAuditListItem_removeMutation.graphql";
import type {
  CompliancePortalAuditListItem_updateVisibilityMutation,
} from "#/__generated__/core/CompliancePortalAuditListItem_updateVisibilityMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const compliancePortalFragment = graphql`
  fragment CompliancePortalAuditListItem_compliancePortal on CompliancePortal {
    id
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

const auditFragment = graphql`
  fragment CompliancePortalAuditListItem_audit on Audit
  @argumentDefinitions(compliancePortalId: { type: "ID!" }) {
    id
    name
    framework {
      name
    }
    validity {
      end
    }
    state
    compliancePortalAudit(compliancePortalId: $compliancePortalId) {
      id
      visibility
    }
  }
`;

const updateAuditVisibilityMutation = graphql`
  mutation CompliancePortalAuditListItem_updateVisibilityMutation(
    $input: UpdateCompliancePortalAuditVisibilityInput!
    $compliancePortalId: ID!
  ) {
    updateCompliancePortalAuditVisibility(input: $input) {
      catalogAudit {
        id
        visibility
      }
      audit {
        id
        compliancePortalAudit(compliancePortalId: $compliancePortalId) {
          id
          visibility
        }
      }
    }
  }
`;

const removeAuditMutation = graphql`
  mutation CompliancePortalAuditListItem_removeMutation(
    $input: DeleteCompliancePortalAuditInput!
    $compliancePortalId: ID!
  ) {
    deleteCompliancePortalAudit(input: $input) {
      deletedCompliancePortalAuditId @deleteRecord
      audit {
        id
        compliancePortalAudit(compliancePortalId: $compliancePortalId) {
          id
        }
      }
    }
  }
`;

export function CompliancePortalAuditListItem(props: {
  compliancePortalKey: CompliancePortalAuditListItem_compliancePortal$key;
  auditKey: CompliancePortalAuditListItem_audit$key;
}) {
  const organizationId = useOrganizationId();
  const { i18n, t } = useTranslation("organizations/compliance-portals");
  const visibilityOptions = getCompliancePortalLinkedVisibilityOptions(t);

  const compliancePortal = useFragment<CompliancePortalAuditListItem_compliancePortal$key>(
    compliancePortalFragment,
    props.compliancePortalKey,
  );
  const audit = useFragment<CompliancePortalAuditListItem_audit$key>(
    auditFragment,
    props.auditKey,
  );
  const catalogAudit = audit.compliancePortalAudit;
  const serverLinked = catalogAudit !== null;
  const [pendingLinked, setPendingLinked] = useState<boolean | null>(null);
  const isLinked = pendingLinked ?? serverLinked;

  const [updateAuditVisibility, isUpdatingAuditVisibility]
    = useMutation<CompliancePortalAuditListItem_updateVisibilityMutation>(
      updateAuditVisibilityMutation,
      {
        successMessage: t("auditListItem.messages.visibilityUpdated"),
        errorToast: t("auditListItem.errors.updateVisibility"),
      },
    );

  const [removeAudit, isRemoving]
    = useMutation<CompliancePortalAuditListItem_removeMutation>(
      removeAuditMutation,
      {
        successMessage: t("auditListItem.messages.removed"),
        errorToast: t("auditListItem.errors.remove"),
      },
    );

  const handleVisibilityChange = useCallback(
    async (value: string) => {
      if (!catalogAudit) {
        return;
      }

      const typedValue = value === "PUBLIC" ? "PUBLIC" : "RESTRICTED";
      await updateAuditVisibility({
        variables: {
          input: {
            compliancePortalId: compliancePortal.id,
            auditId: audit.id,
            compliancePortalVisibility: typedValue,
          },
          compliancePortalId: compliancePortal.id,
        },
      });
    },
    [audit.id, catalogAudit, compliancePortal.id, updateAuditVisibility],
  );

  const handleLinkedChange = useCallback(
    async (checked: boolean) => {
      if (!compliancePortal.canUpdate || checked === isLinked) {
        return;
      }

      setPendingLinked(checked);

      try {
        if (checked) {
          await updateAuditVisibility({
            variables: {
              input: {
                compliancePortalId: compliancePortal.id,
                auditId: audit.id,
                compliancePortalVisibility: "RESTRICTED",
              },
              compliancePortalId: compliancePortal.id,
            },
          });
          setPendingLinked(null);
          return;
        }

        if (!catalogAudit) {
          setPendingLinked(null);
          return;
        }

        await removeAudit({
          variables: {
            input: {
              id: catalogAudit.id,
            },
            compliancePortalId: compliancePortal.id,
          },
        });
        setPendingLinked(null);
      } catch {
        setPendingLinked(null);
      }
    },
    [
      catalogAudit,
      compliancePortal.canUpdate,
      compliancePortal.id,
      isLinked,
      removeAudit,
      updateAuditVisibility,
      audit.id,
    ],
  );

  const isMutating = isUpdatingAuditVisibility || isRemoving;
  const auditTitle = audit.name || t("auditListItem.untitled");
  const validUntilFormatted = audit.validity?.end
    ? dateFormat(i18n.language, audit.validity.end)
    : t("auditListItem.noExpiry");

  return (
    <Tr to={`/organizations/${organizationId}/governance/audits/${audit.id}`}>
      <Td noLink>
        <Checkbox
          checked={isLinked}
          onChange={checked => void handleLinkedChange(checked)}
          disabled={isMutating || !compliancePortal.canUpdate}
          aria-label={t("auditListItem.actions.toggle", {
            title: auditTitle,
          })}
        />
      </Td>
      <Td>
        <div className="flex gap-4 items-center">{audit.framework?.name}</div>
      </Td>
      <Td>{auditTitle}</Td>
      <Td>{validUntilFormatted}</Td>
      <Td>
        <Badge variant={getAuditStateVariant(audit.state)}>
          {t(`auditListItem.states.${audit.state.toLowerCase()}`)}
        </Badge>
      </Td>
      <Td noLink width={130}>
        <Field
          type="select"
          value={catalogAudit?.visibility ?? "RESTRICTED"}
          onValueChange={value => void handleVisibilityChange(value)}
          disabled={!isLinked || isMutating || !compliancePortal.canUpdate}
          className="w-26.25"
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
    </Tr>
  );
}
