import { useOutletContext } from "react-router";

import type { TrustCenterGraphQuery$data } from "/__generated__/core/TrustCenterGraphQuery.graphql";
import { TrustCenterReferencesSection } from "/components/trustCenter/TrustCenterReferencesSection";

export default function TrustCenterReferencesTab() {
  const { organization } = useOutletContext<TrustCenterGraphQuery$data>();

  return (
    <div className="space-y-4">
      {organization.trustCenter?.id && (
        <TrustCenterReferencesSection
          trustCenterId={organization.trustCenter.id}
          canCreateReference={!!organization.trustCenter.canCreateReference}
        />
      )}
    </div>
  );
}
