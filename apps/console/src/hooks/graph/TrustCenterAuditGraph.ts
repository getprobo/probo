import { graphql } from "relay-runtime";
import { useTranslate } from "@probo/i18n";

import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type { TrustCenterAuditGraphUpdateMutation } from "/__generated__/core/TrustCenterAuditGraphUpdateMutation.graphql";

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
      errorMessage: __("Failed to update audit visibility"),
    },
  );
}
