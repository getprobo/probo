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

import { Button, IconChevronDown, IconPlusLarge, IconTrashCan, InternalControlBadge, Spinner } from "@probo/ui";
import { useRef } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, usePaginationFragment } from "react-relay";
import { Link } from "react-router";

import type { TreatmentPlanInternalControlList_internalControl$key } from "#/__generated__/core/TreatmentPlanInternalControlList_internalControl.graphql";
import type { TreatmentPlanInternalControlList_meta$key } from "#/__generated__/core/TreatmentPlanInternalControlList_meta.graphql";
import type { TreatmentPlanInternalControlList_treatmentPlan$key } from "#/__generated__/core/TreatmentPlanInternalControlList_treatmentPlan.graphql";
import type { TreatmentPlanInternalControlListCreateMutation } from "#/__generated__/core/TreatmentPlanInternalControlListCreateMutation.graphql";
import type { TreatmentPlanInternalControlListDetachMutation } from "#/__generated__/core/TreatmentPlanInternalControlListDetachMutation.graphql";
import type { TreatmentPlanInternalControlListPaginationQuery } from "#/__generated__/core/TreatmentPlanInternalControlListPaginationQuery.graphql";
import { LinkedInternalControlDialog } from "#/components/internal-controls/LinkedInternalControlsDialog";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import InternalControlFormDialog from "#/pages/organizations/internal-controls/dialog/InternalControlFormDialog";

const PAGE_SIZE = 50;

export const treatmentPlanInternalControlListMetaFragment = graphql`
  fragment TreatmentPlanInternalControlList_meta on TreatmentPlan {
    id
    treatment
    canUpdate: permission(action: "risk-management:treatment-plan:update")
    organization {
      canCreateInternalControl: permission(action: "core:internal-control:create")
    }
  }
`;

export const treatmentPlanInternalControlListFragment = graphql`
  fragment TreatmentPlanInternalControlList_treatmentPlan on TreatmentPlan
  @refetchable(queryName: "TreatmentPlanInternalControlListPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
    asOf: { type: "Datetime", defaultValue: null }
  ) {
    id
    internalControls(first: $first, after: $after, asOf: $asOf)
      @connection(key: "TreatmentPlanInternalControlList_internalControls", filters: ["asOf"]) {
      __id
      asOf
      edges {
        node {
          id
          ...TreatmentPlanInternalControlList_internalControl
        }
      }
    }
  }
`;

const internalControlFragment = graphql`
  fragment TreatmentPlanInternalControlList_internalControl on InternalControl {
    id
    name
    state
  }
`;

const attachInternalControlMutation = graphql`
  mutation TreatmentPlanInternalControlListCreateMutation(
    $input: CreateTreatmentPlanInternalControlMappingInput!
    $connections: [ID!]!
  ) {
    createTreatmentPlanInternalControlMapping(input: $input) {
      internalControlEdge @prependEdge(connections: $connections) {
        node {
          id
          ...TreatmentPlanInternalControlList_internalControl
        }
      }
      treatmentPlanEdge {
        node {
          id
          netLikelihood
          netImpact
          netRiskScore
          internalControlCount: internalControls(first: 0) {
            totalCount
          }
          implementedInternalControls: internalControls(
            first: 0
            filter: { state: IMPLEMENTED }
          ) {
            totalCount
          }
          inProgressInternalControls: internalControls(
            first: 0
            filter: { state: IN_PROGRESS }
          ) {
            totalCount
          }
          notImplementedInternalControls: internalControls(
            first: 0
            filter: { state: NOT_IMPLEMENTED }
          ) {
            totalCount
          }
        }
      }
    }
  }
`;

const detachInternalControlMutation = graphql`
  mutation TreatmentPlanInternalControlListDetachMutation(
    $input: DeleteTreatmentPlanInternalControlMappingInput!
    $connections: [ID!]!
  ) {
    deleteTreatmentPlanInternalControlMapping(input: $input) {
      deletedInternalControlId @deleteEdge(connections: $connections)
      deletedTreatmentPlanId
      treatmentPlan {
        id
        netLikelihood
        netImpact
        netRiskScore
        internalControlCount: internalControls(first: 0) {
          totalCount
        }
        implementedInternalControls: internalControls(first: 0, filter: { state: IMPLEMENTED }) {
          totalCount
        }
        inProgressInternalControls: internalControls(first: 0, filter: { state: IN_PROGRESS }) {
          totalCount
        }
        notImplementedInternalControls: internalControls(
          first: 0
          filter: { state: NOT_IMPLEMENTED }
        ) {
          totalCount
        }
      }
    }
  }
`;

interface TreatmentPlanInternalControlListProps {
  treatmentPlanKey: TreatmentPlanInternalControlList_treatmentPlan$key & TreatmentPlanInternalControlList_meta$key;
  onChanged?: () => void;
}

export function TreatmentPlanInternalControlList({
  treatmentPlanKey,
  onChanged,
}: TreatmentPlanInternalControlListProps) {
  const { t } = useTranslation();
  const meta = useFragment(
    treatmentPlanInternalControlListMetaFragment,
    treatmentPlanKey as TreatmentPlanInternalControlList_meta$key,
  );
  const { data: treatmentPlan, hasNext, isLoadingNext, loadNext }
    = usePaginationFragment<
      TreatmentPlanInternalControlListPaginationQuery,
      TreatmentPlanInternalControlList_treatmentPlan$key
    >(
      treatmentPlanInternalControlListFragment,
      treatmentPlanKey as TreatmentPlanInternalControlList_treatmentPlan$key,
    );
  const internalControls = treatmentPlan.internalControls?.edges?.map(edge => edge.node) ?? [];
  const connectionId = treatmentPlan.internalControls?.__id ?? "";
  const [attachInternalControl, isAttaching] = useMutation<TreatmentPlanInternalControlListCreateMutation>(
    attachInternalControlMutation,
  );
  const [detachInternalControl, isDetaching] = useMutation<TreatmentPlanInternalControlListDetachMutation>(
    detachInternalControlMutation,
  );
  const isLoading = isAttaching || isDetaching;
  const inFlightRef = useRef(false);
  const accepted = meta.treatment === "ACCEPTED";
  const hasAsOf = treatmentPlan.internalControls?.asOf != null;
  const readOnly = !meta.canUpdate || accepted || hasAsOf;
  const canCreateInternalControl = meta.organization.canCreateInternalControl;

  const onAttach = async (internalControlId: string) => {
    if (inFlightRef.current || isLoading) {
      return;
    }

    inFlightRef.current = true;
    try {
      await attachInternalControl({
        variables: {
          input: {
            internalControlId,
            treatmentPlanId: treatmentPlan.id,
          },
          connections: [connectionId],
        },
      });
      onChanged?.();
    } finally {
      inFlightRef.current = false;
    }
  };

  const onDetach = async (internalControlId: string) => {
    if (inFlightRef.current || isLoading) {
      return;
    }

    inFlightRef.current = true;
    try {
      await detachInternalControl({
        variables: {
          input: {
            internalControlId,
            treatmentPlanId: treatmentPlan.id,
          },
          connections: [connectionId],
        },
      });
      onChanged?.();
    } finally {
      inFlightRef.current = false;
    }
  };

  return (
    <div className="space-y-2" onClick={event => event.stopPropagation()}>
      <div className="flex items-center justify-between gap-3">
        <h3 className="text-sm font-medium text-txt-primary">
          {t("treatmentPlanInternalControlList.title")}
        </h3>
        {!readOnly && (
          <div className="flex items-center gap-3">
            {canCreateInternalControl && (
              <InternalControlFormDialog onCreated={onAttach}>
                <button
                  type="button"
                  disabled={isLoading}
                  className="flex cursor-pointer items-center gap-1 text-sm text-txt-secondary hover:text-txt-primary disabled:opacity-60"
                >
                  <IconPlusLarge size={16} />
                  {t("treatmentPlanInternalControlList.actions.create")}
                </button>
              </InternalControlFormDialog>
            )}
            <LinkedInternalControlDialog
              connectionId={connectionId}
              disabled={isLoading}
              linkedInternalControls={internalControls}
              onLink={(internalControlId) => {
                void onAttach(internalControlId);
              }}
              onUnlink={(internalControlId) => {
                void onDetach(internalControlId);
              }}
            >
              <button
                type="button"
                disabled={isLoading}
                className="flex cursor-pointer items-center gap-1 text-sm text-txt-secondary hover:text-txt-primary disabled:opacity-60"
              >
                <IconPlusLarge size={16} />
                {t("treatmentPlanInternalControlList.actions.link")}
              </button>
            </LinkedInternalControlDialog>
          </div>
        )}
      </div>
      {internalControls.length === 0 && (
        <p className="text-sm text-txt-secondary">
          {accepted
            ? t("treatmentPlanInternalControlList.acceptedEmpty")
            : t("treatmentPlanInternalControlList.empty")}
        </p>
      )}
      {internalControls.length > 0 && (
        <ul
          className={
            readOnly
              ? "grid w-full grid-cols-[minmax(0,1fr)_auto] items-center gap-x-8 divide-y divide-border-low"
              : "grid w-full grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-x-8 divide-y divide-border-low"
          }
        >
          {internalControls.map(internalControl => (
            <InternalControlRow
              key={internalControl.id}
              internalControlKey={internalControl}
              readOnly={readOnly}
              hasAsOf={hasAsOf}
              disabled={isLoading}
              onDetach={(internalControlId) => {
                void onDetach(internalControlId);
              }}
            />
          ))}
        </ul>
      )}
      {hasNext && (
        <Button
          variant="tertiary"
          className="mx-auto"
          disabled={isLoadingNext}
          icon={isLoadingNext ? Spinner : IconChevronDown}
          onClick={() => loadNext(PAGE_SIZE)}
        >
          {t("treatmentPlanInternalControlList.actions.showMore")}
        </Button>
      )}
    </div>
  );
}

function InternalControlRow({
  internalControlKey,
  readOnly,
  hasAsOf,
  disabled,
  onDetach,
}: {
  internalControlKey: TreatmentPlanInternalControlList_internalControl$key & { id: string };
  readOnly: boolean;
  hasAsOf: boolean;
  disabled: boolean;
  onDetach: (internalControlId: string) => void;
}) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const internalControl = useFragment(internalControlFragment, internalControlKey);
  const deleted = internalControl.name === "";

  return (
    <li className="col-span-full grid grid-cols-subgrid items-center py-2.5">
      {deleted
        ? (
            <span className="min-w-0 truncate text-sm text-txt-secondary">
              {t("treatmentPlanInternalControlList.deleted")}
            </span>
          )
        : hasAsOf
          ? (
              <span className="min-w-0 truncate text-sm text-txt-primary">
                {internalControl.name}
              </span>
            )
          : (
              <Link
                to={`/organizations/${organizationId}/governance/internal-controls/${internalControl.id}`}
                className="min-w-0 truncate text-sm text-txt-primary hover:underline"
              >
                {internalControl.name}
              </Link>
            )}
      <InternalControlBadge state={internalControl.state} />
      {!readOnly && (
        <button
          type="button"
          disabled={disabled}
          className="cursor-pointer justify-self-end p-0.5 text-txt-tertiary hover:text-txt-primary disabled:pointer-events-none disabled:opacity-60"
          aria-label={t("treatmentPlanInternalControlList.actions.unlink")}
          onClick={() => onDetach(internalControl.id)}
        >
          <IconTrashCan size={16} />
        </button>
      )}
    </li>
  );
}
