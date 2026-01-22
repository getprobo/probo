import { faviconUrl, formatDate } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Avatar,
  Button,
  DropdownItem,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  RiskBadge,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import {
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { useParams } from "react-router";

import type { VendorGraphListQuery } from "#/__generated__/core/VendorGraphListQuery.graphql";
import type {
  VendorGraphPaginatedFragment$data,
  VendorGraphPaginatedFragment$key,
} from "#/__generated__/core/VendorGraphPaginatedFragment.graphql";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import {
  paginatedVendorsFragment,
  useDeleteVendor,
  vendorsQuery,
} from "#/hooks/graph/VendorGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import { CreateVendorDialog } from "./dialogs/CreateVendorDialog";

type Vendor = NodeOf<VendorGraphPaginatedFragment$data["vendors"]>;

type Props = {
  queryRef: PreloadedQuery<VendorGraphListQuery>;
};

export default function VendorsPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  const data = usePreloadedQuery(vendorsQuery, props.queryRef);
  const pagination = usePaginationFragment(
    paginatedVendorsFragment,
    data.node as VendorGraphPaginatedFragment$key,
  );

  const vendors = pagination.data.vendors?.edges.map(edge => edge.node);
  const connectionId = pagination.data.vendors.__id;

  usePageTitle(__("Vendors"));

  const hasAnyAction
    = !isSnapshotMode
      && vendors.some(({ canUpdate, canDelete }) => canUpdate || canDelete);

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
      <PageHeader
        title={__("Vendors")}
        description={__(
          "Vendors are third-party services that your company uses. Add them to keep track of their risk and compliance status.",
        )}
      >
        {!isSnapshotMode && data.node.canCreateVendor && (
          <CreateVendorDialog
            connection={connectionId}
            organizationId={organizationId}
          >
            <Button icon={IconPlusLarge}>{__("Add vendor")}</Button>
          </CreateVendorDialog>
        )}
      </PageHeader>
      <SortableTable {...pagination}>
        <Thead>
          <Tr>
            <SortableTh field="NAME">{__("Vendor")}</SortableTh>
            <Th>{__("Accessed At")}</Th>
            <Th>{__("Data Risk")}</Th>
            <Th>{__("Business Risk")}</Th>
            {hasAnyAction && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {vendors?.map(vendor => (
            <VendorRow
              key={vendor.id}
              vendor={vendor}
              organizationId={organizationId}
              connectionId={connectionId}
              hasAnyAction={hasAnyAction}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}

function VendorRow({
  vendor,
  organizationId,
  connectionId,
  hasAnyAction,
}: {
  vendor: Vendor;
  organizationId: string;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const { __ } = useTranslate();
  const latestAssessment = vendor.riskAssessments?.edges[0]?.node;
  const deleteVendor = useDeleteVendor(vendor, connectionId);

  const vendorUrl
    = isSnapshotMode && snapshotId
      ? `/organizations/${organizationId}/snapshots/${snapshotId}/vendors/${vendor.id}/overview`
      : `/organizations/${organizationId}/vendors/${vendor.id}/overview`;

  return (
    <>
      <Tr to={vendorUrl}>
        <Td>
          <div className="flex gap-2 items-center">
            <Avatar name={vendor.name} src={faviconUrl(vendor.websiteUrl)} />
            <div>{vendor.name}</div>
          </div>
        </Td>
        <Td>
          {latestAssessment?.createdAt
            ? formatDate(latestAssessment.createdAt)
            : __("Not assessed")}
        </Td>
        <Td>
          <RiskBadge level={latestAssessment?.dataSensitivity ?? "NONE"} />
        </Td>
        <Td>
          <RiskBadge level={latestAssessment?.businessImpact ?? "NONE"} />
        </Td>
        {hasAnyAction && (
          <Td noLink width={50} className="text-end">
            <ActionDropdown>
              {vendor.canDelete && (
                <DropdownItem
                  onClick={deleteVendor}
                  variant="danger"
                  icon={IconTrashCan}
                >
                  {__("Delete")}
                </DropdownItem>
              )}
            </ActionDropdown>
          </Td>
        )}
      </Tr>
    </>
  );
}
