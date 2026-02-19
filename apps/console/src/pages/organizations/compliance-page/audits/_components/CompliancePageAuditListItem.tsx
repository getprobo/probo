import { formatDate, getReportStateLabel, getReportStateVariant, getTrustCenterVisibilityOptions } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Field, Option, Td, Tr } from "@probo/ui";
import { useCallback } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageAuditListItem_compliancePageFragment$key } from "#/__generated__/core/CompliancePageAuditListItem_compliancePageFragment.graphql";
import type { CompliancePageAuditListItem_reportFragment$key } from "#/__generated__/core/CompliancePageAuditListItem_reportFragment.graphql";
import type { CompliancePageAuditListItem_updateReportVisibilityMutation } from "#/__generated__/core/CompliancePageAuditListItem_updateReportVisibilityMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const compliancePageFragment = graphql`
  fragment CompliancePageAuditListItem_compliancePageFragment on TrustCenter {
    canUpdate: permission(action: "core:trust-center:update")
  }
`;

const reportFragment = graphql`
  fragment CompliancePageAuditListItem_reportFragment on Report {
    id
    name
    frameworkType
    framework {
      name
    }
    validUntil
    state
    trustCenterVisibility
  }
`;

const updateReportVisibilityMutation = graphql`
  mutation CompliancePageAuditListItem_updateReportVisibilityMutation($input: UpdateReportInput!) {
    updateReport(input: $input) {
      report {
        ...CompliancePageAuditListItem_reportFragment
      }
    }
  }
`;

export function CompliancePageAuditListItem(props: {
  reportFragmentRef: CompliancePageAuditListItem_reportFragment$key;
  compliancePageFragmentRef: CompliancePageAuditListItem_compliancePageFragment$key;
}) {
  const { reportFragmentRef, compliancePageFragmentRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const compliancePage = useFragment<CompliancePageAuditListItem_compliancePageFragment$key>(
    compliancePageFragment,
    compliancePageFragmentRef,
  );
  const report = useFragment<CompliancePageAuditListItem_reportFragment$key>(reportFragment, reportFragmentRef);

  const [updateReportVisibility, isUpdatingReportVisibility] = useMutationWithToasts<
    CompliancePageAuditListItem_updateReportVisibilityMutation
  >(
    updateReportVisibilityMutation,
    {
      successMessage: __("Report visibility updated successfully."),
      errorMessage: __("Failed to update report visibility"),
    },
  );
  const handleVisibilityChange = useCallback(
    async (value: string) => {
      const stringValue = typeof value === "string" ? value : "";
      const typedValue = stringValue as "NONE" | "PRIVATE" | "PUBLIC";
      await updateReportVisibility({
        variables: {
          input: {
            id: report.id,
            trustCenterVisibility: typedValue,
          },
        },
      });
    },
    [report.id, updateReportVisibility],
  );

  const visibilityOptions = getTrustCenterVisibilityOptions(__);
  const validUntilFormatted = report.validUntil
    ? formatDate(report.validUntil)
    : __("No expiry");

  return (
    <Tr to={`/organizations/${organizationId}/reports/${report.id}`}>
      <Td>
        <div className="flex gap-1 items-baseline">
          <span>{report.framework.name}</span>
          {report.frameworkType && (
            <span className="text-sm italic text-neutral-500">{report.frameworkType}</span>
          )}
        </div>
      </Td>
      <Td>{report.name || __("Untitled")}</Td>
      <Td>{validUntilFormatted}</Td>
      <Td>
        <Badge variant={getReportStateVariant(report.state)}>
          {getReportStateLabel(__, report.state)}
        </Badge>
      </Td>
      <Td noLink width={130} className="pr-0">
        <Field
          type="select"
          value={report.trustCenterVisibility}
          onValueChange={value => void handleVisibilityChange(value)}
          disabled={isUpdatingReportVisibility || !compliancePage.canUpdate}
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
    </Tr>
  );
}
