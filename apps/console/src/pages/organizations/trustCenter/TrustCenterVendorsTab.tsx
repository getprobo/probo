import { Spinner } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useOutletContext } from "react-router";
import { TrustCenterVendorsCard } from "/components/trustCenter/TrustCenterVendorsCard";
import { useTrustCenterVendorUpdate } from "/hooks/graph/TrustCenterVendorGraph";
import type { TrustCenterVendorsCardFragment$key } from "/__generated__/core/TrustCenterVendorsCardFragment.graphql";

type ContextType = {
  organization: {
    vendors?: {
      edges: Array<{
        node: TrustCenterVendorsCardFragment$key;
      }>;
    };
  };
};

export default function TrustCenterVendorsTab() {
  const { __ } = useTranslate();
  const { organization } = useOutletContext<ContextType>();
  const [updateVendorVisibility, isUpdatingVendors] =
    useTrustCenterVendorUpdate();

  const vendors = organization.vendors?.edges?.map((edge) => edge.node) || [];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("Vendors")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__("Manage vendor assessments and third-party risk information")}
          </p>
        </div>
        {isUpdatingVendors && <Spinner />}
      </div>
      <TrustCenterVendorsCard
        vendors={vendors}
        params={{}}
        disabled={isUpdatingVendors}
        onToggleVisibility={updateVendorVisibility}
      />
    </div>
  );
}
