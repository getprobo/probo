import { faviconUrl, validateSnapshotConsistency } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
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
import {
  ConnectionHandler,
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import { Outlet, useParams } from "react-router";

import type { VendorComplianceTabFragment$key } from "#/__generated__/core/VendorComplianceTabFragment.graphql";
import type { VendorGraphNodeQuery } from "#/__generated__/core/VendorGraphNodeQuery.graphql";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import {
  useDeleteVendor,
  vendorConnectionKey,
  vendorNodeQuery,
} from "#/hooks/graph/VendorGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { ImportAssessmentDialog } from "./dialogs/ImportAssessmentDialog";
import { complianceReportsFragment } from "./tabs/VendorComplianceTab";

type Props = {
  queryRef: PreloadedQuery<VendorGraphNodeQuery>;
};

export default function VendorDetailPage(props: Props) {
  const { node: vendor } = usePreloadedQuery(vendorNodeQuery, props.queryRef);
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  validateSnapshotConsistency(vendor, snapshotId);
  const deleteVendor = useDeleteVendor(
    vendor,
    ConnectionHandler.getConnectionID(organizationId, vendorConnectionKey),
  );
  const logo = faviconUrl(vendor.websiteUrl);
  const reportsCount = useFragment(
    complianceReportsFragment,
    vendor as VendorComplianceTabFragment$key,
  ).complianceReports.edges.length;

  const vendorsUrl
    = isSnapshotMode && snapshotId
      ? `/organizations/${organizationId}/snapshots/${snapshotId}/vendors`
      : `/organizations/${organizationId}/vendors`;

  const baseVendorUrl
    = isSnapshotMode && snapshotId
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
            {vendor.canAssess && (
              <ImportAssessmentDialog vendorId={vendor.id}>
                <Button icon={IconPageTextLine} variant="secondary">
                  {__("Assessment From Website")}
                </Button>
              </ImportAssessmentDialog>
            )}
            {vendor.canDelete && (
              <ActionDropdown variant="secondary">
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  onClick={deleteVendor}
                >
                  {__("Delete")}
                </DropdownItem>
              </ActionDropdown>
            )}
          </div>
        )}
      </div>

      <Tabs>
        <TabLink to={`${baseVendorUrl}/overview`}>{__("Overview")}</TabLink>
        <TabLink to={`${baseVendorUrl}/certifications`}>
          {__("Certifications")}
        </TabLink>
        <TabLink to={`${baseVendorUrl}/compliance`}>
          {__("Compliance reports")}
          {reportsCount > 0 && <TabBadge>{reportsCount}</TabBadge>}
        </TabLink>
        <TabLink to={`${baseVendorUrl}/risks`}>{__("Risk Assessment")}</TabLink>
        <TabLink to={`${baseVendorUrl}/contacts`}>{__("Contacts")}</TabLink>
        <TabLink to={`${baseVendorUrl}/services`}>{__("Services")}</TabLink>
      </Tabs>

      <Outlet context={{ vendor }} />
    </div>
  );
}
