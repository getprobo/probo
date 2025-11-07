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
import { Outlet, useParams } from "react-router";
import { faviconUrl, validateSnapshotConsistency } from "@probo/helpers";
import { ImportAssessmentDialog } from "./dialogs/ImportAssessmentDialog";
import { complianceReportsFragment } from "./tabs/VendorComplianceTab";
import type { VendorComplianceTabFragment$key } from "./tabs/__generated__/VendorComplianceTabFragment.graphql";
import { SnapshotBanner } from "/components/SnapshotBanner";
import { Authorized } from "/permissions";

type Props = {
  queryRef: PreloadedQuery<VendorGraphNodeQuery>;
};

export default function VendorDetailPage(props: Props) {
  const data = usePreloadedQuery(vendorNodeQuery, props.queryRef);
  const vendor = data.node;
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  if (!vendor) {
    return <div>{__("Vendor not found")}</div>;
  }

  validateSnapshotConsistency(vendor, snapshotId);
  const deleteVendor = useDeleteVendor(
    vendor,
    ConnectionHandler.getConnectionID(organizationId, vendorConnectionKey)
  );
  const logo = faviconUrl(vendor.websiteUrl);
  const reportsCount = useFragment(
    complianceReportsFragment,
    vendor as VendorComplianceTabFragment$key
  ).complianceReports.edges.length;

  const vendorsUrl = isSnapshotMode && snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/vendors`
    : `/organizations/${organizationId}/vendors`;

  const baseVendorUrl = isSnapshotMode && snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/vendors/${vendor.id}`
    : `/organizations/${organizationId}/vendors/${vendor.id}`;

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
      <Breadcrumb
        items={[
          {
            label: __("Vendors"),
            to: vendorsUrl,
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
        {!isSnapshotMode && (
          <div className="flex gap-2 items-center">
            <ImportAssessmentDialog vendorId={vendor.id!}>
              <Button icon={IconPageTextLine} variant="secondary">
                {__("Assessment From Website")}
              </Button>
            </ImportAssessmentDialog>
            <Authorized entity="Vendor" action="deleteVendor">
              <ActionDropdown variant="secondary">
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  onClick={deleteVendor}
                >
                  {__("Delete")}
                </DropdownItem>
              </ActionDropdown>
            </Authorized>
          </div>
        )}
      </div>

      <Tabs>
        <TabLink to={`${baseVendorUrl}/overview`}>
          {__("Overview")}
        </TabLink>
        <TabLink to={`${baseVendorUrl}/certifications`}>
          {__("Certifications")}
        </TabLink>
        <TabLink to={`${baseVendorUrl}/compliance`}>
          {__("Compliance reports")}
          {reportsCount > 0 && <TabBadge>{reportsCount}</TabBadge>}
        </TabLink>
        <TabLink to={`${baseVendorUrl}/risks`}>
          {__("Risk Assessment")}
        </TabLink>
        <TabLink to={`${baseVendorUrl}/contacts`}>
          {__("Contacts")}
        </TabLink>
        <TabLink to={`${baseVendorUrl}/services`}>
          {__("Services")}
        </TabLink>
      </Tabs>

      <Outlet context={{ vendor }} />
    </div>
  );
}
