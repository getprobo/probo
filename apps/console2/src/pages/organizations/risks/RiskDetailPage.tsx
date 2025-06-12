import {
  ActionDropdown,
  Avatar,
  Badge,
  Breadcrumb,
  Button,
  Drawer,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  PageHeader,
  PropertyRow,
  TabBadge,
  TabLink,
  Tabs,
  useConfirm,
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
  const confirm = useConfirm();

  const onDelete = () => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      RisksConnectionKey
    );
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteRisk({
            variables: {
              input: { riskId },
              connections: [connectionId],
            },
            onSuccess() {
              navigate(`/organizations/${organizationId}/risks`);
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete the risk "%s". This action cannot be undone.'
          ),
          risk.name
        ),
      }
    );
  };

  const documentsCount = risk.documentsInfo?.totalCount ?? 0;
  const measuresCount = risk.measuresInfo?.totalCount ?? 0;
  const controlsCount = risk.controlsInfo?.totalCount ?? 0;

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
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={onDelete}
            >
              {__("Delete")}
            </DropdownItem>
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
          <TabBadge>{measuresCount}</TabBadge>
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/risks/${riskId}/documents`}
        >
          {__("Documents")}
          <TabBadge>{documentsCount}</TabBadge>
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/risks/${riskId}/controls`}
        >
          {__("Controls")}
          <TabBadge>{controlsCount}</TabBadge>
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
