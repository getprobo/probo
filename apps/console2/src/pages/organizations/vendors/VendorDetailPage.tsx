import {
  ConnectionHandler,
  useFragment,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import type { VendorGraphNodeQuery } from "/hooks/graph/__generated__/VendorGraphNodeQuery.graphql";
import {
  useDeleteVendor,
  vendorConnectionKey,
  vendorNodeQuery,
} from "/hooks/graph/VendorGraph";
import {
  ActionDropdown,
  Breadcrumb,
  Button,
  DropdownItem,
  IconPageTextLine,
  IconTrashCan,
  TabBadge,
  TabLink,
  Tabs,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { Outlet } from "react-router";
import { faviconUrl } from "@probo/helpers";
import { ImportAssessmentDialog } from "./dialogs/ImportAssessmentDialog";
import { complianceReportsFragment } from "./tabs/VendorComplianceTab";
import type { VendorComplianceTabFragment$key } from "./tabs/__generated__/VendorComplianceTabFragment.graphql";

type Props = {
  queryRef: PreloadedQuery<VendorGraphNodeQuery>;
};

export default function VendorDetailPage(props: Props) {
  const data = usePreloadedQuery(vendorNodeQuery, props.queryRef);
  const vendor = data.node;
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const deleteVendor = useDeleteVendor(
    vendor,
    ConnectionHandler.getConnectionID(organizationId, vendorConnectionKey)
  );
  const logo = faviconUrl(vendor.websiteUrl);
  const reportsCount = useFragment(
    complianceReportsFragment,
    vendor as VendorComplianceTabFragment$key
  ).complianceReports.edges.length;

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: __("Vendors"),
            to: `/organizations/${organizationId}/vendors`,
          },
          {
            label: vendor.name ?? "",
          },
        ]}
      />
      <div className="flex justify-between items-start">
        <div className="space-y-4">
          {logo && (
            <img
              src={logo}
              alt={vendor.name ?? ""}
              className="shadow-mid rounded-2xl"
            />
          )}
          <div className="text-2xl">{vendor.name}</div>
        </div>
        <div className="flex gap-2 items-center">
          <ImportAssessmentDialog vendorId={vendor.id!}>
            <Button icon={IconPageTextLine} variant="secondary">
              {__("Assessment From Website")}
            </Button>
          </ImportAssessmentDialog>
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={deleteVendor}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        </div>
      </div>

      <Tabs>
        <TabLink
          to={`/organizations/${organizationId}/vendors/${vendor.id}/overview`}
        >
          {__("Overview")}
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/vendors/${vendor.id}/certifications`}
        >
          {__("Certifications")}
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/vendors/${vendor.id}/compliance`}
        >
          {__("Compliance reports")}
          {reportsCount > 0 && <TabBadge>{reportsCount}</TabBadge>}
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/vendors/${vendor.id}/risks`}
        >
          {__("Risk Assessment")}
        </TabLink>
      </Tabs>

      <Outlet context={{ vendor, peopleId: data.viewer.user.people!.id }} />
    </div>
  );
}
