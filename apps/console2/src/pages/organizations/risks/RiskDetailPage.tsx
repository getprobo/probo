import {
  ActionDropdown,
  Avatar,
  Badge,
  Breadcrumb,
  Button,
  ConfirmDialog,
  Drawer,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  PageHeader,
  PropertyRow,
  TabLink,
  Tabs,
} from "@probo/ui";
import { Outlet, useNavigate, useParams } from "react-router";
import { useTranslate } from "@probo/i18n";
import { getTreatment, sprintf } from "@probo/helpers";
import { ConnectionHandler } from "relay-runtime";
import { usePreloadedQuery, type PreloadedQuery } from "react-relay";
import FormRiskDialog from "./FormRiskDialog";
import { usePageTitle } from "@probo/hooks";
import { useOrganizationId } from "/hooks/useOrganizationId";
import {
  riskNodeQuery,
  RisksConnectionKey,
  useDeleteRiskMutation,
} from "/hooks/graph/RiskGraph";
import type { RiskGraphNodeQuery } from "/hooks/graph/__generated__/RiskGraphNodeQuery.graphql";

type Props = {
  queryRef: PreloadedQuery<RiskGraphNodeQuery>;
};

export default function RiskDetailPage(props: Props) {
  const { riskId } = useParams<{ riskId: string }>();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  if (!riskId) {
    throw new Error("Cannot load risk detail page without riskId parameter");
  }

  const { __ } = useTranslate();
  const data = usePreloadedQuery(riskNodeQuery, props.queryRef);
  const risk = data.node;
  const [deleteRisk] = useDeleteRiskMutation();

  usePageTitle(risk.name ?? "Risk detail");

  const onDelete = (riskId: string) => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      RisksConnectionKey
    );
    deleteRisk({
      variables: {
        input: { riskId },
        connections: [connectionId],
      },
      onSuccess() {
        navigate(`/organizations/${organizationId}/risks`);
      },
    });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center mb-4">
        <Breadcrumb
          items={[
            {
              label: __("Risks"),
              to: `/organizations/${organizationId}/risks`,
            },
            {
              label: __("Risk detail"),
            },
          ]}
        />
        <div className="flex gap-2">
          <FormRiskDialog
            trigger={
              <Button icon={IconPencil} variant="secondary">
                {__("Edit")}
              </Button>
            }
            risk={{ id: riskId, ...risk }}
          />
          <ActionDropdown variant="secondary">
            <ConfirmDialog
              message={sprintf(
                __(
                  'This will permanently delete the risk "%s". This action cannot be undone.'
                ),
                risk.name
              )}
              onConfirm={() => onDelete(riskId)}
            >
              <DropdownItem variant="danger" icon={IconTrashCan}>
                {__("Delete")}
              </DropdownItem>
            </ConfirmDialog>
          </ActionDropdown>
        </div>
      </div>

      <PageHeader title={risk.name} />

      <Tabs>
        <TabLink
          to={`/organizations/${organizationId}/risks/${riskId}/overview`}
        >
          {__("Overview")}
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/risks/${riskId}/measures`}
        >
          {__("Measures")}
        </TabLink>
      </Tabs>

      <Outlet context={{ risk }} />

      <Drawer>
        <PropertyRow label={__("Owner")}>
          <Badge variant="highlight" size="md" className="gap-2">
            <Avatar name={risk.owner?.fullName ?? ""} />
            {risk.owner?.fullName}
          </Badge>
        </PropertyRow>
        <PropertyRow label={__("Treatment")}>
          <Badge variant="highlight" size="md" className="gap-2">
            {getTreatment(__, risk.treatment)}
          </Badge>
        </PropertyRow>
        <PropertyRow label={__("Note")}>
          <div className="text-sm text-txt-secondary">{risk.note}</div>
        </PropertyRow>
      </Drawer>
    </div>
  );
}
