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
  ConfirmDialog,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import { ConnectionHandler, graphql } from "relay-runtime";
import { useLazyLoadQuery } from "react-relay";
import { useOrganizationId } from "../../../hooks/useOrganizationId";
import type {
  VendorsPageQuery,
  VendorsPageQuery$data,
} from "./__generated__/VendorsPageQuery.graphql";
import { faviconUrl, sprintf } from "@probo/helpers";
import type { NodeOf } from "../../../types";
import { CreateVendorDialog } from "./CreateVendorDialog";
import { useDeleteVendorMutation } from "../../../graph/VendorGraph";

const vendorsQuery = graphql`
  query VendorsPageQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        vendors(first: 25) @connection(key: "VendorsPage_vendors") {
          __id
          edges {
            node {
              id
              name
              websiteUrl
              updatedAt
              riskAssessments(
                first: 1
                orderBy: { direction: DESC, field: ASSESSED_AT }
              ) {
                edges {
                  node {
                    id
                    assessedAt
                    expiresAt
                    dataSensitivity
                    businessImpact
                  }
                }
              }
            }
          }
        }
      }
    }
  }
`;

type Vendor = NodeOf<Required<VendorsPageQuery$data["node"]>["vendors"]>;

export default function VendorsPage() {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const data = useLazyLoadQuery<VendorsPageQuery>(vendorsQuery, {
    organizationId,
  });
  console.log(data);

  const vendors = data.node?.vendors?.edges.map((edge) => edge.node);

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
          trigger={<Button icon={IconPlusLarge}>{__("Add vendor")}</Button>}
        />
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
}: {
  vendor: Vendor;
  organizationId: string;
}) {
  const { __, dateFormat } = useTranslate();
  const latestAssessment = vendor.riskAssessments?.edges[0]?.node;
  const isExpired = latestAssessment
    ? new Date(latestAssessment.expiresAt) < new Date()
    : false;

  const [deleteVendor] = useDeleteVendorMutation();

  const onDelete = (vendorId: string) => {
    deleteVendor({
      variables: {
        input: {
          vendorId,
        },
        connections: [
          ConnectionHandler.getConnectionID(
            organizationId,
            "VendorsPage_vendors"
          ),
        ],
      },
    });
  };

  return (
    <Tr to={`/organizations/${organizationId}/vendors/${vendor.id}`}>
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
      <Td noLink>
        <ActionDropdown>
          <ConfirmDialog
            message={sprintf(
              __(
                'This will permanently delete the vendor "%s". This action cannot be undone.'
              ),
              vendor.name
            )}
            onConfirm={() => onDelete(vendor.id)}
          >
            <DropdownItem variant="danger" icon={IconTrashCan}>
              {__("Delete")}
            </DropdownItem>
          </ConfirmDialog>
        </ActionDropdown>
      </Td>
    </Tr>
  );
}
