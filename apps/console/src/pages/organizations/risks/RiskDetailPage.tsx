import {
  getTreatment,
  sprintf,
  validateSnapshotConsistency,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
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
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet, useNavigate, useParams } from "react-router";
import { ConnectionHandler } from "relay-runtime";

import type { RiskGraphNodeQuery } from "#/__generated__/core/RiskGraphNodeQuery.graphql";
import { SnapshotBanner } from "#/components/SnapshotBanner";
import {
  riskNodeQuery,
  RisksConnectionKey,
  useDeleteRiskMutation,
} from "#/hooks/graph/RiskGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import FormRiskDialog from "./FormRiskDialog";

type Props = {
  queryRef: PreloadedQuery<RiskGraphNodeQuery>;
};

export default function RiskDetailPage(props: Props) {
  const { riskId, snapshotId } = useParams<{
    riskId: string;
    snapshotId?: string;
  }>();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const isSnapshotMode = Boolean(snapshotId);

  if (!riskId) {
    throw new Error("Cannot load risk detail page without riskId parameter");
  }

  const { __ } = useTranslate();
  const { node: risk } = usePreloadedQuery<RiskGraphNodeQuery>(
    riskNodeQuery,
    props.queryRef,
  );

  validateSnapshotConsistency(risk, snapshotId);
  const [deleteRisk] = useDeleteRiskMutation();

  usePageTitle(risk.name ?? "Risk detail");
  const confirm = useConfirm();

  const onDelete = () => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      RisksConnectionKey,
      { filter: { snapshotId: snapshotId || null } },
    );
    confirm(
      () =>
        new Promise<void>((resolve) => {
          void deleteRisk({
            variables: {
              input: { riskId },
              connections: [connectionId],
            },
            onSuccess() {
              const risksUrl
                = isSnapshotMode && snapshotId
                  ? `/organizations/${organizationId}/snapshots/${snapshotId}/risks`
                  : `/organizations/${organizationId}/risks`;
              void navigate(risksUrl);
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the risk \"%s\". This action cannot be undone.",
          ),
          risk.name,
        ),
      },
    );
  };

  const documentsCount = risk.documentsInfo?.totalCount ?? 0;
  const measuresCount = risk.measuresInfo?.totalCount ?? 0;
  const controlsCount = risk.controlsInfo?.totalCount ?? 0;
  const obligationsCount = risk.obligationsInfo?.totalCount ?? 0;

  const risksUrl
    = isSnapshotMode && snapshotId
      ? `/organizations/${organizationId}/snapshots/${snapshotId}/risks`
      : `/organizations/${organizationId}/risks`;

  const baseTabUrl
    = isSnapshotMode && snapshotId
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
            {risk.canUpdate && (
              <FormRiskDialog
                trigger={(
                  <Button icon={IconPencil} variant="secondary">
                    {__("Edit")}
                  </Button>
                )}
                risk={{ id: riskId, ...risk }}
              />
            )}
            {risk.canDelete && (
              <ActionDropdown variant="secondary">
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  onClick={onDelete}
                >
                  {__("Delete")}
                </DropdownItem>
              </ActionDropdown>
            )}
          </div>
        )}
      </div>

      <PageHeader title={risk.name} description={risk.description} />
      <Tabs>
        <TabLink to={`${baseTabUrl}/overview`}>{__("Overview")}</TabLink>
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
