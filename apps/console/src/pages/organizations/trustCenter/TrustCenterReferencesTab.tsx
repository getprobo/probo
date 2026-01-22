import { useOutletContext } from "react-router";

import { TrustCenterReferencesSection } from "/components/trustCenter/TrustCenterReferencesSection";
import type { TrustCenterGraphQuery$data } from "/__generated__/core/TrustCenterGraphQuery.graphql";

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
