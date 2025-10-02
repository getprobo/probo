import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useTranslate } from "@probo/i18n";
import type { TrustCenterAuditGraphUpdateMutation } from "./__generated__/TrustCenterAuditGraphUpdateMutation.graphql";

export const trustCenterAuditUpdateMutation = graphql`
  mutation TrustCenterAuditGraphUpdateMutation($input: UpdateAuditInput!) {
    updateAudit(input: $input) {
      audit {
        id
        trustCenterVisibility
        ...TrustCenterAuditsCardFragment
      }
    }
  }
`;

export function useTrustCenterAuditUpdate() {
  const { __ } = useTranslate();

  return useMutationWithToasts<TrustCenterAuditGraphUpdateMutation>(
    trustCenterAuditUpdateMutation,
    {
      successMessage: __("Audit visibility updated successfully."),
      errorMessage: __("Failed to update audit visibility. Please try again."),
    }
  );
}
