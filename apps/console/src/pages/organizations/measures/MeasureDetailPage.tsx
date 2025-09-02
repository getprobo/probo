import { Outlet, useNavigate, useParams } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import {
  ActionDropdown,
  Breadcrumb,
  Button,
  Drawer,
  DropdownItem,
  IconCheckmark1,
  IconFrame2,
  IconPageTextLine,
  IconPencil,
  IconTrashCan,
  IconWarning,
  Option,
  PageHeader,
  PropertyRow,
  Select,
  TabBadge,
  TabLink,
  Tabs,
  useConfirm,
} from "@probo/ui";
import { MeasureBadge } from "@probo/ui/src/Molecules/Badge/MeasureBadge";
import { useTranslate } from "@probo/i18n";
import {
  ConnectionHandler,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import type { MeasureGraphNodeQuery } from "/hooks/graph/__generated__/MeasureGraphNodeQuery.graphql";
import {
  MeasureConnectionKey,
  measureNodeQuery,
  useDeleteMeasureMutation,
  useUpdateMeasure,
} from "/hooks/graph/MeasureGraph";
import {
  getMeasureStateLabel,
  measureStates,
  slugify,
  sprintf,
  validateSnapshotConsistency,
} from "@probo/helpers";
import MeasureFormDialog from "./dialog/MeasureFormDialog";
import { SnapshotBanner } from "/components/SnapshotBanner";

type Props = {
  queryRef: PreloadedQuery<MeasureGraphNodeQuery>;
};

export default function MeasureDetailPage(props: Props) {
  const { measureId, snapshotId } = useParams<{ measureId: string; snapshotId?: string }>();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const isSnapshotMode = Boolean(snapshotId);

  if (!measureId) {
    throw new Error(
      "Cannot load measure detail page without measureId parameter",
    );
  }

  const data = usePreloadedQuery(measureNodeQuery, props.queryRef);
  const measure = data.node;
  const { __ } = useTranslate();
  const [deleteMeasure] = useDeleteMeasureMutation();
  const confirm = useConfirm();
  const [updateMeasure, isUpdating] = useUpdateMeasure();

  if (isSnapshotMode) {
    validateSnapshotConsistency(measure, snapshotId);
  }

  const tasksCount = measure.tasksInfos?.totalCount ?? 0;
  const evidencesCount = measure.evidencesInfos?.totalCount ?? 0;
  const controlsCount = measure.controlsInfos?.totalCount ?? 0;
  const risksCount = measure.risksInfos?.totalCount ?? 0;

  const measuresUrl = isSnapshotMode && snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/risks`
    : `/organizations/${organizationId}/measures`;

  const baseTabUrl = isSnapshotMode && snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/risks/measures/${measureId}`
    : `/organizations/${organizationId}/measures/${measureId}`;

  const onDelete = () => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      MeasureConnectionKey,
    );
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteMeasure({
            variables: {
              input: { measureId },
              connections: [connectionId],
            },
            onSuccess() {
              navigate(`/organizations/${organizationId}/measures`);
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete the measure "%s". This action cannot be undone.',
          ),
          measure.name,
        ),
      },
    );
  };

  const onStateChange = (state: string) => {
    updateMeasure({
      variables: {
        input: {
          id: measureId,
          state,
        },
      },
    });
  };

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}

      {/* Header */}
      <Breadcrumb
        items={[
          {
            label: isSnapshotMode ? __("Risks") : __("Measures"),
            to: measuresUrl,
          },
          ...(measure.category && !isSnapshotMode
            ? [
                {
                  label: measure.category,
                  to: `/organizations/${organizationId}/measures/category/${slugify(measure.category)}`,
                },
              ]
            : []),
          {
            label: __("Measure detail"),
          },
        ]}
      />

      <PageHeader title={measure.name} description={measure.description}>
        {!isSnapshotMode && (
          <>
            <MeasureFormDialog measure={measure}>
              <Button variant="secondary" icon={IconPencil}>
                {__("Edit")}
              </Button>
            </MeasureFormDialog>
            <Select
              disabled={isUpdating}
              onValueChange={onStateChange}
              name="state"
              placeholder={__("Select state")}
              className="rounded-full"
              value={measure.state}
            >
              {measureStates.map((state) => (
                <Option key={state} value={state}>
                  {getMeasureStateLabel(__, state)}
                </Option>
              ))}
            </Select>
            <ActionDropdown variant="secondary">
              <DropdownItem variant="danger" icon={IconTrashCan} onClick={onDelete}>
                {__("Delete")}
              </DropdownItem>
            </ActionDropdown>
          </>
        )}
      </PageHeader>

      <Tabs>
        <TabLink to={`${baseTabUrl}/evidences`}>
          <IconPageTextLine size={20} />
          {__("Evidences")}
          <TabBadge>{evidencesCount}</TabBadge>
        </TabLink>
        {!isSnapshotMode && (
          <>
            <TabLink to={`${baseTabUrl}/tasks`}>
              <IconCheckmark1 size={20} />
              {__("Tasks")}
              <TabBadge>{tasksCount}</TabBadge>
            </TabLink>
            <TabLink to={`${baseTabUrl}/controls`}>
              <IconFrame2 size={20} />
              {__("Controls")}
              <TabBadge>{controlsCount}</TabBadge>
            </TabLink>
            <TabLink to={`${baseTabUrl}/risks`}>
              <IconWarning size={20} />
              {__("Risks")}
              <TabBadge>{risksCount}</TabBadge>
            </TabLink>
          </>
        )}
      </Tabs>

      <Outlet context={{ measure }} />

      <Drawer>
        <PropertyRow label={__("State")}>
          <MeasureBadge state={measure.state!} />
        </PropertyRow>
        <PropertyRow label={__("Category")}>
          <div className="text-sm text-txt-secondary">{measure.category}</div>
        </PropertyRow>
      </Drawer>
    </div>
  );
}
