import { useOutletContext } from "react-router";
import { TrustCenterReferencesSection } from "/components/trustCenter/TrustCenterReferencesSection";

type ContextType = {
  organization: {
    trustCenter?: {
      id: string;
    } | null;
  };
};

export default function TrustCenterReferencesTab() {
  const { organization } = useOutletContext<ContextType>();

  return (
    <div className="space-y-4">
      {organization.trustCenter?.id && (
        <TrustCenterReferencesSection trustCenterId={organization.trustCenter.id} />
      )}
    </div>
  );
}
