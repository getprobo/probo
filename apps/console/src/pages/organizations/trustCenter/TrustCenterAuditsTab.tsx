import { Spinner } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useOutletContext } from "react-router";
import { TrustCenterAuditsCard } from "/components/trustCenter/TrustCenterAuditsCard";
import { useTrustCenterAuditUpdate } from "/hooks/graph/TrustCenterAuditGraph";
import type { TrustCenterAuditsCardFragment$key } from "/__generated__/core/TrustCenterAuditsCardFragment.graphql";

type ContextType = {
  organization: {
    audits?: {
      edges: Array<{
        node: TrustCenterAuditsCardFragment$key;
      }>;
    };
  };
};

export default function TrustCenterAuditsTab() {
  const { __ } = useTranslate();
  const { organization } = useOutletContext<ContextType>();
  const [updateAuditVisibility, isUpdatingAudits] = useTrustCenterAuditUpdate();

  const audits = (organization.audits?.edges ?? []).map((edge) => edge.node);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("Audits")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__("Manage audit reports and compliance certifications")}
          </p>
        </div>
        {isUpdatingAudits && <Spinner />}
      </div>
      <TrustCenterAuditsCard
        audits={audits}
        params={{}}
        disabled={isUpdatingAudits}
        onChangeVisibility={updateAuditVisibility}
      />
    </div>
  );
}
