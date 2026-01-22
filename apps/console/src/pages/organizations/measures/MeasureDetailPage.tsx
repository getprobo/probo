import {
  getMeasureStateLabel,
  measureStates,
  slugify,
  sprintf,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
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
import { Suspense } from "react";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useLazyLoadQuery,
  usePreloadedQuery,
} from "react-relay";
import { Outlet, useNavigate, useParams } from "react-router";

import type { MeasureDetailPageTasksCountQuery } from "#/__generated__/core/MeasureDetailPageTasksCountQuery.graphql";
import type { MeasureGraphNodeQuery } from "#/__generated__/core/MeasureGraphNodeQuery.graphql";
import {
  MeasureConnectionKey,
  measureNodeQuery,
  useDeleteMeasureMutation,
  useUpdateMeasure,
} from "#/hooks/graph/MeasureGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import MeasureFormDialog from "./dialog/MeasureFormDialog";

const tasksCountQuery = graphql`
  query MeasureDetailPageTasksCountQuery($measureId: ID!) {
    node(id: $measureId) {
      ... on Measure {
        tasks(first: 0) {
          totalCount
        }
      }
    }
  }
`;

function TasksCountBadge({ measureId }: { measureId: string }) {
  const data = useLazyLoadQuery<MeasureDetailPageTasksCountQuery>(
    tasksCountQuery,
    { measureId },
  );
  const count = data.node?.tasks?.totalCount ?? 0;
  return <TabBadge>{count}</TabBadge>;
}

type Props = {
  queryRef: PreloadedQuery<MeasureGraphNodeQuery>;
};

export default function MeasureDetailPage(props: Props) {
  const { measureId } = useParams<{ measureId: string }>();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery(measureNodeQuery, props.queryRef);
  const measure = data.node;
  const { __ } = useTranslate();
  const [deleteMeasure] = useDeleteMeasureMutation();
  const navigate = useNavigate();
  const confirm = useConfirm();
  const [updateMeasure, isUpdating] = useUpdateMeasure();
  if (!measureId) {
    throw new Error(
      "Cannot load measure detail page without measureId parameter",
    );
  }

  const evidencesCount = measure.evidencesInfos?.totalCount ?? 0;
  const controlsCount = measure.controlsInfos?.totalCount ?? 0;
  const risksCount = measure.risksInfos?.totalCount ?? 0;

  const onDelete = () => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      MeasureConnectionKey,
    );
    confirm(
      () =>
        new Promise<void>((resolve) => {
          void deleteMeasure({
            variables: {
              input: { measureId },
              connections: [connectionId],
            },
            onSuccess() {
              void navigate(`/organizations/${organizationId}/measures`);
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the measure \"%s\". This action cannot be undone.",
          ),
          measure.name,
        ),
      },
    );
  };

  const onStateChange = (state: string) => {
    void updateMeasure({
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
      {/* Header */}
      <Breadcrumb
        items={[
          {
            label: __("Measures"),
            to: `/organizations/${organizationId}/measures`,
          },
          ...(measure.category
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
        {measure.canUpdate && (
          <>
            <MeasureFormDialog measure={measure}>
              <Button variant="secondary" icon={IconPencil}>
                {__("Edit")}
              </Button>
            </MeasureFormDialog>
            <Select
              disabled={isUpdating}
              onValueChange={state => void onStateChange(state)}
              name="state"
              placeholder={__("Select state")}
              className="rounded-full"
              value={measure.state}
            >
              {measureStates.map(state => (
                <Option key={state} value={state}>
                  {getMeasureStateLabel(__, state)}
                </Option>
              ))}
            </Select>
          </>
        )}
        {measure.canDelete && (
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
      </PageHeader>

      <Tabs>
        <TabLink
          to={`/organizations/${organizationId}/measures/${measureId}/evidences`}
        >
          <IconPageTextLine size={20} />
          {__("Evidences")}
          <TabBadge>{evidencesCount}</TabBadge>
        </TabLink>
        {measure.canListTasks && (
          <TabLink
            to={`/organizations/${organizationId}/measures/${measureId}/tasks`}
          >
            <IconCheckmark1 size={20} />
            {__("Tasks")}
            <Suspense fallback={<TabBadge>-</TabBadge>}>
              <TasksCountBadge measureId={measureId} />
            </Suspense>
          </TabLink>
        )}
        <TabLink
          to={`/organizations/${organizationId}/measures/${measureId}/controls`}
        >
          <IconFrame2 size={20} />
          {__("Controls")}
          <TabBadge>{controlsCount}</TabBadge>
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/measures/${measureId}/risks`}
        >
          <IconWarning size={20} />
          {__("Risks")}
          <TabBadge>{risksCount}</TabBadge>
        </TabLink>
      </Tabs>

      <Outlet context={{ measure }} />

      <Drawer>
        <PropertyRow label={__("State")}>
          <MeasureBadge state={measure.state!} />
        </PropertyRow>
      </Drawer>
    </div>
  );
}
