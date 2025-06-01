import {
  Button,
  IconPlusLarge,
  PageHeader,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Avatar,
  RiskBadge,
  Badge,
  ActionDropdown,
  DropdownItem,
  IconTrashCan,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import { usePreloadedQuery, type PreloadedQuery } from "react-relay";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { faviconUrl } from "@probo/helpers";
import type { NodeOf } from "/types";
import { CreateVendorDialog } from "./CreateVendorDialog";
import { useDeleteVendor, vendorsQuery } from "/hooks/graph/VendorGraph";
import type {
  VendorGraphListQuery,
  VendorGraphListQuery$data,
} from "/hooks/graph/__generated__/VendorGraphListQuery.graphql";

type Vendor = NodeOf<Required<VendorGraphListQuery$data["node"]>["vendors"]>;

type Props = {
  queryRef: PreloadedQuery<VendorGraphListQuery>;
};

export default function VendorsPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const data = usePreloadedQuery(vendorsQuery, props.queryRef);

  const vendors = data.node?.vendors?.edges.map((edge) => edge.node);
  const connectionId = data.node!.vendors!.__id;

  usePageTitle(__("Vendors"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Vendors")}
        description={__(
          "Vendors are third-party services that your company uses. Add them to keep track of their risk and compliance status."
        )}
      >
        <CreateVendorDialog
          connection={data.node!.vendors!.__id}
          organizationId={organizationId}
        >
          <Button icon={IconPlusLarge}>{__("Add vendor")}</Button>
        </CreateVendorDialog>
      </PageHeader>
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Vendor")}</Th>
            <Th>{__("Accessed At")}</Th>
            <Th>{__("Data Risk")}</Th>
            <Th>{__("Business Risk")}</Th>
            <Th>{__("Compliance status")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {vendors?.map((vendor) => (
            <VendorRow
              key={vendor.id}
              vendor={vendor}
              organizationId={organizationId}
              connectionId={connectionId}
            />
          ))}
        </Tbody>
      </Table>
    </div>
  );
}

function VendorRow({
  vendor,
  organizationId,
  connectionId,
}: {
  vendor: Vendor;
  organizationId: string;
  connectionId: string;
}) {
  const { __, dateFormat } = useTranslate();
  const latestAssessment = vendor.riskAssessments?.edges[0]?.node;
  const isExpired = latestAssessment
    ? new Date(latestAssessment.expiresAt) < new Date()
    : false;

  const deleteVendor = useDeleteVendor(vendor, connectionId);

  return (
    <>
      <Tr to={`/organizations/${organizationId}/vendors/${vendor.id}/overview`}>
        <Td>
          <div className="flex gap-2 items-center">
            <Avatar name={vendor.name} src={faviconUrl(vendor.websiteUrl)} />
            <div>{vendor.name}</div>
          </div>
        </Td>
        <Td>
          {latestAssessment?.assessedAt
            ? dateFormat(latestAssessment.assessedAt, {
                day: "2-digit",
                weekday: "short",
                month: "short",
              })
            : __("Not assessed")}
        </Td>
        <Td>
          <RiskBadge level={latestAssessment?.dataSensitivity ?? "NONE"} />
        </Td>
        <Td>
          <RiskBadge level={latestAssessment?.businessImpact ?? "NONE"} />
        </Td>
        <Td>
          <Badge variant={isExpired ? "danger" : "warning"}>
            {isExpired ? __("Late") : __("In progress")}
          </Badge>
        </Td>
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            <DropdownItem
              onClick={deleteVendor}
              variant="danger"
              icon={IconTrashCan}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        </Td>
      </Tr>
    </>
  );
}
