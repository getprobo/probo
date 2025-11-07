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
import { getTreatment, sprintf, validateSnapshotConsistency } from "@probo/helpers";
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
import { SnapshotBanner } from "/components/SnapshotBanner";
import { Authorized } from "/permissions";

type Props = {
  queryRef: PreloadedQuery<RiskGraphNodeQuery>;
};

export default function RiskDetailPage(props: Props) {
  const { riskId, snapshotId } = useParams<{ riskId: string; snapshotId?: string }>();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const isSnapshotMode = Boolean(snapshotId);

  if (!riskId) {
    throw new Error("Cannot load risk detail page without riskId parameter");
  }

  const { __ } = useTranslate();
  const data = usePreloadedQuery<RiskGraphNodeQuery>(riskNodeQuery, props.queryRef);
  const risk = data.node;

  validateSnapshotConsistency(risk, snapshotId);
  const [deleteRisk] = useDeleteRiskMutation();

  usePageTitle(risk.name ?? "Risk detail");
  const confirm = useConfirm();

  const onDelete = () => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      RisksConnectionKey,
      { filter: { snapshotId: snapshotId || null } }
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
              const risksUrl = isSnapshotMode && snapshotId
                ? `/organizations/${organizationId}/snapshots/${snapshotId}/risks`
                : `/organizations/${organizationId}/risks`;
              navigate(risksUrl);
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
  const obligationsCount = risk.obligationsInfo?.totalCount ?? 0;

  const risksUrl = isSnapshotMode && snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/risks`
    : `/organizations/${organizationId}/risks`;

  const baseTabUrl = isSnapshotMode && snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/risks/${riskId}`
    : `/organizations/${organizationId}/risks/${riskId}`;

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
      {/* Header */}
      <div className="flex justify-between items-center mb-4">
        <Breadcrumb
          items={[
            {
              label: __("Risks"),
              to: risksUrl,
            },
            {
              label: __("Risk detail"),
            },
          ]}
        />
        {!isSnapshotMode && (
          <div className="flex gap-2">
            <Authorized entity="Risk" action="updateRisk">
              <FormRiskDialog
                trigger={
                  <Button icon={IconPencil} variant="secondary">
                    {__("Edit")}
                  </Button>
                }
                risk={{ id: riskId, ...risk }}
              />
            </Authorized>
            <Authorized entity="Risk" action="deleteRisk">
              <ActionDropdown variant="secondary">
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  onClick={onDelete}
                >
                  {__("Delete")}
                </DropdownItem>
              </ActionDropdown>
            </Authorized>
          </div>
        )}
      </div>

      <PageHeader title={risk.name} description={risk.description} />
      <Tabs>
        <TabLink to={`${baseTabUrl}/overview`}>
          {__("Overview")}
        </TabLink>
        {!isSnapshotMode && (
          <>
            <TabLink to={`${baseTabUrl}/measures`}>
              {__("Measures")}
              <TabBadge>{measuresCount}</TabBadge>
            </TabLink>
            <TabLink to={`${baseTabUrl}/documents`}>
              {__("Documents")}
              <TabBadge>{documentsCount}</TabBadge>
            </TabLink>
            <TabLink to={`${baseTabUrl}/controls`}>
              {__("Controls")}
              <TabBadge>{controlsCount}</TabBadge>
            </TabLink>
            <TabLink to={`${baseTabUrl}/obligations`}>
              {__("Obligations")}
              <TabBadge>{obligationsCount}</TabBadge>
            </TabLink>
          </>
        )}
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
        <PropertyRow label={__("Initial Risk Score")}>
          <div className="text-sm text-txt-secondary">
            {risk.inherentRiskScore}
          </div>
        </PropertyRow>
        <PropertyRow label={__("Residual Risk Score")}>
          <div className="text-sm text-txt-secondary">
            {risk.residualRiskScore}
          </div>
        </PropertyRow>
        <PropertyRow label={__("Note")}>
          <div className="text-sm text-txt-secondary">{risk.note}</div>
        </PropertyRow>
      </Drawer>
    </div>
  );
}
