import { Outlet, useParams } from "react-router";
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
  MeasureBadge,
  Option,
  PropertyRow,
  Select,
  TabBadge,
  TabLink,
  Tabs,
  useConfirm,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import {
  ConnectionHandler,
  useFragment,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import type { MeasureGraphNodeQuery } from "/hooks/graph/__generated__/MeasureGraphNodeQuery.graphql";
import {
  MeasureConnectionKey,
  measureNodeQuery,
  useDeleteMeasureMutation,
  useUpdateMeasure,
} from "/hooks/graph/MeasureGraph";
import { PageHeader } from "@probo/ui";
import { getMeasureStateLabel, measureStates, slugify } from "@probo/helpers";
import MeasureFormDialog from "./dialog/MeasureFormDialog";
import { sprintf } from "@probo/helpers";
import { useNavigate } from "react-router";
import { tasksFragment } from "./tabs/MeasureTasksTab";
import type { MeasureTasksTabFragment$key } from "./tabs/__generated__/MeasureTasksTabFragment.graphql";
import { evidencesFragment } from "./tabs/MeasureEvidencesTab";
import type { MeasureEvidencesTabFragment$key } from "./tabs/__generated__/MeasureEvidencesTabFragment.graphql";
import { controlsFragment } from "./tabs/MeasureControlsTab";
import type { MeasureControlsTabFragment$key } from "./tabs/__generated__/MeasureControlsTabFragment.graphql";
import { risksFragment } from "./tabs/MeasureRisksTab";
import type { MeasureRisksTabFragment$key } from "./tabs/__generated__/MeasureRisksTabFragment.graphql";

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
      "Cannot load measure detail page without measureId parameter"
    );
  }

  const tasksCount = useFragment(
    tasksFragment,
    measure as MeasureTasksTabFragment$key
  ).tasks.edges.length;
  const evidencesCount = useFragment(
    evidencesFragment,
    measure as MeasureEvidencesTabFragment$key
  ).evidences.edges.length;
  const controlsCount = useFragment(
    controlsFragment,
    measure as MeasureControlsTabFragment$key
  ).controls.edges.length;
  const risksCount = useFragment(
    risksFragment,
    measure as MeasureRisksTabFragment$key
  ).risks.edges.length;

  const onDelete = () => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      MeasureConnectionKey
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
            'This will permanently delete the measure "%s". This action cannot be undone.'
          ),
          measure.name
        ),
      }
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
      </PageHeader>

      <Tabs>
        <TabLink
          to={`/organizations/${organizationId}/measures/${measureId}/evidences`}
        >
          <IconPageTextLine size={20} />
          {__("Evidences")}
          {evidencesCount > 0 && <TabBadge>{evidencesCount}</TabBadge>}
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/measures/${measureId}/tasks`}
        >
          <IconCheckmark1 size={20} />
          {__("Tasks")}
          {tasksCount > 0 && <TabBadge>{tasksCount}</TabBadge>}
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/measures/${measureId}/controls`}
        >
          <IconFrame2 size={20} />
          {__("Controls")}
          {controlsCount > 0 && <TabBadge>{controlsCount}</TabBadge>}
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/measures/${measureId}/risks`}
        >
          <IconWarning size={20} />
          {__("Risks")}
          {risksCount > 0 && <TabBadge>{risksCount}</TabBadge>}
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
