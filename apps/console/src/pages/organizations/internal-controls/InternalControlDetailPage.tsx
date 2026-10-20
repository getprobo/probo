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
  internalControlImplementationStatuses,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { dateFormat, formatDuration } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
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
import { InternalControlBadge } from "@probo/ui/src/Molecules/Badge/InternalControlBadge";
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

import type { InternalControlDetailPageNodeQuery } from "#/__generated__/core/InternalControlDetailPageNodeQuery.graphql";
import type { InternalControlDetailPageTasksCountQuery } from "#/__generated__/core/InternalControlDetailPageTasksCountQuery.graphql";
import type { InternalControlImplementationStatus } from "#/__generated__/core/InternalControlGraphUpdateMutation.graphql";
import {
  InternalControlConnectionKey,
  useDeleteInternalControlMutation,
  useUpdateInternalControl,
} from "#/hooks/graph/InternalControlGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import InternalControlFormDialog from "./dialog/InternalControlFormDialog";
import { controlsFragment } from "./tabs/InternalControlControlsTab";
import { documentsFragment } from "./tabs/InternalControlDocumentsTab";
import { evidencesFragment } from "./tabs/InternalControlEvidencesTab";
import { risksFragment } from "./tabs/InternalControlRisksTab";
import { thirdPartiesFragment } from "./third-parties/InternalControlThirdPartiesPage";

void controlsFragment;
void documentsFragment;
void evidencesFragment;
void risksFragment;
void thirdPartiesFragment;

export const internalControlNodeQuery = graphql`
  query InternalControlDetailPageNodeQuery($internalControlId: ID!) {
    node(id: $internalControlId) {
      ... on InternalControl {
        name
        description
        state
        code
        implementationStatus
        operatingFrequency {
          mode
          interval
          event
        }
        evidenceCadence
        testingCadence
        nextEvidenceDue
        nextTestDue
        owner {
          fullName
        }
        reviewer {
          fullName
        }
        canUpdate: permission(action: "core:internal-control:update")
        canDelete: permission(action: "core:internal-control:delete")
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
        ...InternalControlRisksTabFragment
        ...InternalControlControlsTabFragment
        ...InternalControlDocumentsTabFragment
        ...InternalControlFormDialogInternalControlFragment
        ...InternalControlEvidencesTabFragment
        ...InternalControlThirdPartiesPageFragment
      }
    }
  }
`;

const tasksCountQuery = graphql`
  query InternalControlDetailPageTasksCountQuery($internalControlId: ID!) {
    node(id: $internalControlId) {
      ... on InternalControl {
        tasks(first: 0) {
          totalCount
        }
      }
    }
  }
`;

function TasksCountBadge({ internalControlId }: { internalControlId: string }) {
  const data = useLazyLoadQuery<InternalControlDetailPageTasksCountQuery>(
    tasksCountQuery,
    { internalControlId },
  );
  const count = data.node?.tasks?.totalCount ?? 0;
  return <TabBadge>{count}</TabBadge>;
}

function ReadOnlyImplementationStatus({
  state,
  implementationStatus,
}: {
  state?: string | null;
  implementationStatus?: string | null;
}) {
  const { t } = useTranslation();

  if (state === "NOT_STARTED" || state === "NOT_APPLICABLE" || state === "UNKNOWN") {
    return <InternalControlBadge state={state} />;
  }

  if (implementationStatus === "OPERATING") {
    return (
      <Badge variant="success">
        {t("internalControlDetailPage.implementationStatuses.operating")}
      </Badge>
    );
  }

  if (
    implementationStatus === "NOT_IMPLEMENTED"
    || implementationStatus === "IN_PROGRESS"
    || implementationStatus === "IMPLEMENTED"
  ) {
    return <InternalControlBadge state={implementationStatus} />;
  }

  return null;
}

type Props = {
  queryRef: PreloadedQuery<InternalControlDetailPageNodeQuery>;
};

export default function InternalControlDetailPage(props: Props) {
  const { internalControlId } = useParams<{ internalControlId: string }>();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<InternalControlDetailPageNodeQuery>(internalControlNodeQuery, props.queryRef);
  const internalControl = data.node;
  const { t, i18n } = useTranslation();
  usePageTitle(internalControl.name ?? "");
  const [deleteInternalControl] = useDeleteInternalControlMutation();
  const navigate = useNavigate();
  const confirm = useConfirm();
  const [updateInternalControl, isUpdating] = useUpdateInternalControl();
  if (!internalControlId) {
    throw new Error(
      "Cannot load internalControl detail page without internalControlId parameter",
    );
  }

  const evidencesCount = internalControl.evidencesInfos?.totalCount ?? 0;
  const controlsCount = internalControl.controlsInfos?.totalCount ?? 0;
  const risksCount = internalControl.risksInfos?.totalCount ?? 0;
  const documentsCount = internalControl.documentsInfos?.totalCount ?? 0;
  const thirdPartiesCount = internalControl.thirdPartiesInfos?.totalCount ?? 0;

  const onDelete = () => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      InternalControlConnectionKey,
    );
    confirm(
      () =>
        new Promise<void>((resolve) => {
          void deleteInternalControl({
            variables: {
              input: { internalControlId },
              connections: [connectionId],
            },
            onSuccess() {
              void navigate(`/organizations/${organizationId}/governance/internal-controls`);
              resolve();
            },
          });
        }),
      {
        message: t("internalControlDetailPage.deleteConfirmation", { name: internalControl.name }),
      },
    );
  };

  const onStatusChange = (implementationStatus: InternalControlImplementationStatus) => {
    void updateInternalControl({
      variables: {
        input: {
          id: internalControlId,
          implementationStatus,
        },
      },
    });
  };

  const dueDate = (value?: string | null) => {
    if (!value) {
      return null;
    }

    return dateFormat(i18n.language, value, { dateStyle: "medium" });
  };

  const operating = operatingFrequencyLabel(internalControl.operatingFrequency, t);

  return (
    <div className="space-y-6">
      <PageHeader title={internalControl.name} description={internalControl.description}>
        {!internalControl.canUpdate && (
          <ReadOnlyImplementationStatus
            state={internalControl.state}
            implementationStatus={internalControl.implementationStatus}
          />
        )}
        {internalControl.canUpdate && (
          <>
            <InternalControlFormDialog internalControl={internalControl}>
              <Button variant="secondary" icon={IconPencil}>
                {t("internalControlDetailPage.actions.edit")}
              </Button>
            </InternalControlFormDialog>
            <Select
              disabled={isUpdating}
              onValueChange={(status) => {
                if ((internalControlImplementationStatuses as readonly string[]).includes(status)) {
                  onStatusChange(status as InternalControlImplementationStatus);
                }
              }}
              name="implementationStatus"
              placeholder={t("internalControlDetailPage.fields.selectState")}
              className="rounded-full"
              value={internalControl.implementationStatus ?? undefined}
            >
              {internalControlImplementationStatuses.map(status => (
                <Option key={status} value={status}>
                  {t(`internalControlDetailPage.implementationStatuses.${status.toLowerCase()}`)}
                </Option>
              ))}
            </Select>
          </>
        )}
        {internalControl.canDelete && (
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={onDelete}
            >
              {t("internalControlDetailPage.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </PageHeader>
      <div className="flex flex-wrap gap-x-6 gap-y-1 text-sm text-txt-tertiary">
        {internalControl.code && (
          <span>
            {t("internalControlDetailPage.fields.code")}
            {": "}
            {internalControl.code}
          </span>
        )}
        {internalControl.owner?.fullName && (
          <span>
            {t("internalControlDetailPage.fields.owner")}
            {": "}
            {internalControl.owner.fullName}
          </span>
        )}
        {internalControl.reviewer?.fullName && (
          <span>
            {t("internalControlDetailPage.fields.reviewer")}
            {": "}
            {internalControl.reviewer.fullName}
          </span>
        )}
        {operating && (
          <span>
            {t("internalControlDetailPage.fields.operatingFrequency")}
            {": "}
            {operating}
          </span>
        )}
        {formatDuration(internalControl.evidenceCadence, t) && (
          <span>
            {t("internalControlDetailPage.fields.evidenceCadence")}
            {": "}
            {formatDuration(internalControl.evidenceCadence, t)}
          </span>
        )}
        {dueDate(internalControl.nextEvidenceDue) && (
          <span>
            {t("internalControlDetailPage.fields.nextEvidenceDue")}
            {": "}
            {dueDate(internalControl.nextEvidenceDue)}
          </span>
        )}
        {formatDuration(internalControl.testingCadence, t) && (
          <span>
            {t("internalControlDetailPage.fields.testingCadence")}
            {": "}
            {formatDuration(internalControl.testingCadence, t)}
          </span>
        )}
        {dueDate(internalControl.nextTestDue) && (
          <span>
            {t("internalControlDetailPage.fields.nextTestDue")}
            {": "}
            {dueDate(internalControl.nextTestDue)}
          </span>
        )}
      </div>

      <Tabs>
        <TabLink
          to={`/organizations/${organizationId}/governance/internal-controls/${internalControlId}/evidences`}
        >
          <IconPageCheck size={20} />
          {t("internalControlDetailPage.tabs.evidences")}
          <TabBadge>{evidencesCount}</TabBadge>
        </TabLink>
        {internalControl.canListTasks && (
          <TabLink
            to={`/organizations/${organizationId}/governance/internal-controls/${internalControlId}/tasks`}
          >
            <IconCheckmark1 size={20} />
            {t("internalControlDetailPage.tabs.tasks")}
            <Suspense fallback={<TabBadge>-</TabBadge>}>
              <TasksCountBadge internalControlId={internalControlId} />
            </Suspense>
          </TabLink>
        )}
        <TabLink
          to={`/organizations/${organizationId}/governance/internal-controls/${internalControlId}/controls`}
        >
          <IconFrame2 size={20} />
          {t("internalControlDetailPage.tabs.controls")}
          <TabBadge>{controlsCount}</TabBadge>
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/governance/internal-controls/${internalControlId}/risks`}
        >
          <IconWarning size={20} />
          {t("internalControlDetailPage.tabs.risks")}
          <TabBadge>{risksCount}</TabBadge>
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/governance/internal-controls/${internalControlId}/documents`}
        >
          <IconPageTextLine size={20} />
          {t("internalControlDetailPage.tabs.documents")}
          <TabBadge>{documentsCount}</TabBadge>
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/governance/internal-controls/${internalControlId}/third-parties`}
        >
          <IconStore size={20} />
          {t("internalControlDetailPage.tabs.thirdParties")}
          <TabBadge>{thirdPartiesCount}</TabBadge>
        </TabLink>
      </Tabs>

      <Outlet context={{ internalControl }} />
    </div>
  );
}

function operatingFrequencyLabel(
  frequency: {
    readonly mode: string;
    readonly interval?: string | null;
    readonly event?: string | null;
  } | null | undefined,
  t: (key: string, options?: { count?: number }) => string,
): string | null {
  if (!frequency) {
    return null;
  }

  switch (frequency.mode) {
    case "CONTINUOUS":
      return t("internalControlDetailPage.operatingModes.continuous");
    case "EVENT":
      return frequency.event || t("internalControlDetailPage.operatingModes.event");
    case "PERIODIC":
      return formatDuration(frequency.interval, t);
    default:
      return null;
  }
}
