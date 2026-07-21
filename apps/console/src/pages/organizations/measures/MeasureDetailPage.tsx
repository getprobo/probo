// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import {
  measureStates,
} from "@probo/helpers";
import {
  ActionDropdown,
  Breadcrumb,
  Button,
  DropdownItem,
  IconCheckmark1,
  IconFrame2,
  IconPageCheck,
  IconPageTextLine,
  IconPencil,
  IconStore,
  IconTrashCan,
  IconWarning,
  Option,
  PageHeader,
  Select,
  TabBadge,
  TabLink,
  Tabs,
  useConfirm,
} from "@probo/ui";
import { MeasureBadge } from "@probo/ui/src/Molecules/Badge/MeasureBadge";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useLazyLoadQuery,
  usePreloadedQuery,
} from "react-relay";
import { Outlet, useNavigate, useParams } from "react-router";

import type { MeasureDetailPageNodeQuery } from "#/__generated__/core/MeasureDetailPageNodeQuery.graphql";
import type { MeasureDetailPageTasksCountQuery } from "#/__generated__/core/MeasureDetailPageTasksCountQuery.graphql";
import {
  MeasureConnectionKey,
  useDeleteMeasureMutation,
  useUpdateMeasure,
} from "#/hooks/graph/MeasureGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import MeasureFormDialog from "./dialog/MeasureFormDialog";
import { controlsFragment } from "./tabs/MeasureControlsTab";
import { documentsFragment } from "./tabs/MeasureDocumentsTab";
import { evidencesFragment } from "./tabs/MeasureEvidencesTab";
import { risksFragment } from "./tabs/MeasureRisksTab";
import { thirdPartiesFragment } from "./third-parties/MeasureThirdPartiesPage";

void controlsFragment;
void documentsFragment;
void evidencesFragment;
void risksFragment;
void thirdPartiesFragment;

export const measureNodeQuery = graphql`
  query MeasureDetailPageNodeQuery($measureId: ID!) {
    node(id: $measureId) {
      ... on Measure {
        name
        description
        state
        category
        canUpdate: permission(action: "core:measure:update")
        canDelete: permission(action: "core:measure:delete")
        canListTasks: permission(action: "core:task:list")
        evidencesInfos: evidences(first: 0) {
          totalCount
        }
        risksInfos: risks(first: 0) {
          totalCount
        }
        controlsInfos: controls(first: 0) {
          totalCount
        }
        documentsInfos: documents(first: 0) {
          totalCount
        }
        thirdPartiesInfos: thirdParties(first: 0) {
          totalCount
        }
        ...MeasureRisksTabFragment
        ...MeasureControlsTabFragment
        ...MeasureDocumentsTabFragment
        ...MeasureFormDialogMeasureFragment
        ...MeasureEvidencesTabFragment
        ...MeasureThirdPartiesPageFragment
      }
    }
  }
`;

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
  queryRef: PreloadedQuery<MeasureDetailPageNodeQuery>;
};

export default function MeasureDetailPage(props: Props) {
  const { measureId } = useParams<{ measureId: string }>();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<MeasureDetailPageNodeQuery>(measureNodeQuery, props.queryRef);
  const measure = data.node;
  const { t } = useTranslation();
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
  const documentsCount = measure.documentsInfos?.totalCount ?? 0;
  const thirdPartiesCount = measure.thirdPartiesInfos?.totalCount ?? 0;

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
        message: t("measureDetailPage.deleteConfirmation", { name: measure.name }),
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
            label: t("measureDetailPage.breadcrumb.measures"),
            to: `/organizations/${organizationId}/measures`,
          },
          ...(measure.category
            ? [
                {
                  label: measure.category,
                  to: `/organizations/${organizationId}/measures?category=${encodeURIComponent(measure.category)}`,
                },
              ]
            : []),
          {
            label: t("measureDetailPage.breadcrumb.detail"),
          },
        ]}
      />

      <PageHeader title={measure.name} description={measure.description}>
        {!measure.canUpdate && <MeasureBadge state={measure.state!} />}
        {measure.canUpdate && (
          <>
            <MeasureFormDialog measure={measure}>
              <Button variant="secondary" icon={IconPencil}>
                {t("measureDetailPage.actions.edit")}
              </Button>
            </MeasureFormDialog>
            <Select
              disabled={isUpdating}
              onValueChange={state => void onStateChange(state)}
              name="state"
              placeholder={t("measureDetailPage.fields.selectState")}
              className="rounded-full"
              value={measure.state}
            >
              {measureStates.map(state => (
                <Option key={state} value={state}>
                  {t(`measureDetailPage.states.${state.toLowerCase()}`)}
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
              {t("measureDetailPage.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </PageHeader>

      <Tabs>
        <TabLink
          to={`/organizations/${organizationId}/measures/${measureId}/evidences`}
        >
          <IconPageCheck size={20} />
          {t("measureDetailPage.tabs.evidences")}
          <TabBadge>{evidencesCount}</TabBadge>
        </TabLink>
        {measure.canListTasks && (
          <TabLink
            to={`/organizations/${organizationId}/measures/${measureId}/tasks`}
          >
            <IconCheckmark1 size={20} />
            {t("measureDetailPage.tabs.tasks")}
            <Suspense fallback={<TabBadge>-</TabBadge>}>
              <TasksCountBadge measureId={measureId} />
            </Suspense>
          </TabLink>
        )}
        <TabLink
          to={`/organizations/${organizationId}/measures/${measureId}/controls`}
        >
          <IconFrame2 size={20} />
          {t("measureDetailPage.tabs.controls")}
          <TabBadge>{controlsCount}</TabBadge>
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/measures/${measureId}/risks`}
        >
          <IconWarning size={20} />
          {t("measureDetailPage.tabs.risks")}
          <TabBadge>{risksCount}</TabBadge>
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/measures/${measureId}/documents`}
        >
          <IconPageTextLine size={20} />
          {t("measureDetailPage.tabs.documents")}
          <TabBadge>{documentsCount}</TabBadge>
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/measures/${measureId}/third-parties`}
        >
          <IconStore size={20} />
          {t("measureDetailPage.tabs.thirdParties")}
          <TabBadge>{thirdPartiesCount}</TabBadge>
        </TabLink>
      </Tabs>

      <Outlet context={{ measure }} />
    </div>
  );
}
