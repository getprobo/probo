// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { TrashIcon, XIcon } from "@phosphor-icons/react";
import { dateFormat } from "@probo/i18n";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useRefetchableFragment } from "react-relay";

import type { FindingCorrectiveActionPlanItem_plan$key } from "#/__generated__/core/FindingCorrectiveActionPlanItem_plan.graphql";
import type { FindingCorrectiveActionPlanItemDeleteMutation } from "#/__generated__/core/FindingCorrectiveActionPlanItemDeleteMutation.graphql";
import type { FindingCorrectiveActionPlanItemLinkMutation } from "#/__generated__/core/FindingCorrectiveActionPlanItemLinkMutation.graphql";
import type { FindingCorrectiveActionPlanItemRefetchQuery } from "#/__generated__/core/FindingCorrectiveActionPlanItemRefetchQuery.graphql";
import type { FindingCorrectiveActionPlanItemUnlinkMutation } from "#/__generated__/core/FindingCorrectiveActionPlanItemUnlinkMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { InternalControlCombobox } from "#/pages/organizations/tasks/_components/InternalControlCombobox";
import { taskDetailsPath } from "#/pages/organizations/tasks/_lib/taskPath";

import { correctiveActionPlans, findingsPage } from "./variants";

const planFragment = graphql`
  fragment FindingCorrectiveActionPlanItem_plan on CorrectiveActionPlan
  @refetchable(queryName: "FindingCorrectiveActionPlanItemRefetchQuery") {
    id
    title
    description
    dueDate
    owner {
      id
      fullName
    }
    canDelete: permission(action: "core:corrective-action-plan:delete")
    canLinkInternalControl: permission(action: "core:corrective-action-plan:create-internal-control-mapping")
    canUnlinkInternalControl: permission(action: "core:corrective-action-plan:delete-internal-control-mapping")
    internalControls(first: 50, orderBy: { field: NAME, direction: ASC }) {
      totalCount
      edges {
        node {
          id
          name
        }
      }
    }
    tasks(first: 50) {
      totalCount
      edges {
        node {
          id
          name
        }
      }
    }
  }
`;

const deletePlanMutation = graphql`
  mutation FindingCorrectiveActionPlanItemDeleteMutation(
    $input: DeleteCorrectiveActionPlanInput!
    $connections: [ID!]!
  ) {
    deleteCorrectiveActionPlan(input: $input) {
      deletedCorrectiveActionPlanId @deleteEdge(connections: $connections)
    }
  }
`;

const linkInternalControlMutation = graphql`
  mutation FindingCorrectiveActionPlanItemLinkMutation(
    $input: CreateCorrectiveActionPlanInternalControlMappingInput!
  ) {
    createCorrectiveActionPlanInternalControlMapping(input: $input) {
      correctiveActionPlanEdge {
        node {
          id
        }
      }
    }
  }
`;

const unlinkInternalControlMutation = graphql`
  mutation FindingCorrectiveActionPlanItemUnlinkMutation(
    $input: DeleteCorrectiveActionPlanInternalControlMappingInput!
  ) {
    deleteCorrectiveActionPlanInternalControlMapping(input: $input) {
      deletedCorrectiveActionPlanId
    }
  }
`;

interface FindingCorrectiveActionPlanItemProps {
  planKey: FindingCorrectiveActionPlanItem_plan$key;
  connectionId: string;
}

export function FindingCorrectiveActionPlanItem({
  planKey,
  connectionId,
}: FindingCorrectiveActionPlanItemProps) {
  const { t, i18n } = useTranslation("organizations/findings");
  const organizationId = useOrganizationId();
  const [plan, refetch] = useRefetchableFragment<
    FindingCorrectiveActionPlanItemRefetchQuery,
    FindingCorrectiveActionPlanItem_plan$key
  >(planFragment, planKey);
  const [deletePlan, isDeleting] = useMutation<FindingCorrectiveActionPlanItemDeleteMutation>(
    deletePlanMutation,
    { errorToast: t("details.plans.deleteError") },
  );
  const [linkInternalControl, isLinking] = useMutation<FindingCorrectiveActionPlanItemLinkMutation>(
    linkInternalControlMutation,
    { errorToast: t("details.plans.linkError") },
  );
  const [unlinkInternalControl, isUnlinking] = useMutation<FindingCorrectiveActionPlanItemUnlinkMutation>(
    unlinkInternalControlMutation,
    { errorToast: t("details.plans.unlinkError") },
  );
  const { links, linkRow } = findingsPage();
  const { plan: planSlot, header, meta, description, group } = correctiveActionPlans();
  const internalControls = plan.internalControls.edges.map(edge => edge.node);
  const tasks = plan.tasks.edges.map(edge => edge.node);
  const linkedIds = new Set(internalControls.map(internalControl => internalControl.id));
  const busy = isDeleting || isLinking || isUnlinking;

  function reload() {
    refetch({}, { fetchPolicy: "network-only" });
  }

  const metaParts = [
    plan.owner?.fullName,
    plan.dueDate
      ? t("details.plans.due", { date: dateFormat(i18n.language, plan.dueDate, { dateStyle: "medium" }) })
      : null,
  ].filter(part => part != null && part !== "");

  return (
    <div className={planSlot()}>
      <div className={header()}>
        <Text size={2} weight="medium">{plan.title}</Text>
        {plan.canDelete && (
          <IconButton
            size={1}
            variant="ghost"
            color="neutral"
            aria-label={t("details.plans.delete")}
            disabled={busy}
            onClick={() => {
              void deletePlan({
                variables: {
                  input: { correctiveActionPlanId: plan.id },
                  connections: [connectionId],
                },
              }).catch(() => {
                // The mutation hook already reported the failure.
              });
            }}
          >
            <TrashIcon />
          </IconButton>
        )}
      </div>
      {metaParts.length > 0 && (
        <Text size={1} color="faint" className={meta()}>{metaParts.join(" · ")}</Text>
      )}
      {plan.description && (
        <Text size={1} className={description()}>{plan.description}</Text>
      )}
      <div className={group()}>
        <Text size={1} weight="medium">{t("details.plans.internalControls")}</Text>
        {internalControls.length === 0
          ? <Text size={1} color="faint">{t("details.plans.noInternalControls")}</Text>
          : (
              <div className={links()}>
                {internalControls.map(internalControl => (
                  <div key={internalControl.id} className={linkRow()}>
                    <Link
                      size={1}
                      to={`/organizations/${organizationId}/governance/internal-controls/${internalControl.id}`}
                    >
                      {internalControl.name}
                    </Link>
                    {plan.canUnlinkInternalControl && (
                      <IconButton
                        size={1}
                        variant="ghost"
                        color="neutral"
                        aria-label={t("details.plans.unlinkInternalControl")}
                        disabled={busy}
                        onClick={() => {
                          void unlinkInternalControl({
                            variables: {
                              input: {
                                correctiveActionPlanId: plan.id,
                                internalControlId: internalControl.id,
                              },
                            },
                          }).then(reload).catch(() => {
                            // The mutation hook already reported the failure.
                          });
                        }}
                      >
                        <XIcon />
                      </IconButton>
                    )}
                  </div>
                ))}
              </div>
            )}
        {plan.canLinkInternalControl && (
          <InternalControlCombobox
            excludeIds={linkedIds}
            disabled={busy}
            onSelect={(internalControl) => {
              void linkInternalControl({
                variables: {
                  input: {
                    correctiveActionPlanId: plan.id,
                    internalControlId: internalControl.id,
                  },
                },
              }).then(reload).catch(() => {
                // The mutation hook already reported the failure.
              });
            }}
          />
        )}
      </div>
      <div className={group()}>
        <Text size={1} weight="medium">
          {t("details.plans.tasks", { count: plan.tasks.totalCount })}
        </Text>
        {tasks.length === 0
          ? <Text size={1} color="faint">{t("details.plans.noTasks")}</Text>
          : (
              <div className={links()}>
                {tasks.map(task => (
                  <Link key={task.id} size={1} to={taskDetailsPath(organizationId, task.id)}>
                    {task.name}
                  </Link>
                ))}
              </div>
            )}
      </div>
    </div>
  );
}
