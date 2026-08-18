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

import { Button, IconChevronDown, IconPlusLarge, IconTrashCan, MeasureBadge, Spinner } from "@probo/ui";
import { useRef } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, usePaginationFragment } from "react-relay";
import { Link } from "react-router";

import type { TreatmentPlanMeasureList_measure$key } from "#/__generated__/core/TreatmentPlanMeasureList_measure.graphql";
import type { TreatmentPlanMeasureList_treatmentPlan$key } from "#/__generated__/core/TreatmentPlanMeasureList_treatmentPlan.graphql";
import type { TreatmentPlanMeasureListCreateMutation } from "#/__generated__/core/TreatmentPlanMeasureListCreateMutation.graphql";
import type { TreatmentPlanMeasureListDetachMutation } from "#/__generated__/core/TreatmentPlanMeasureListDetachMutation.graphql";
import type { TreatmentPlanMeasureListPaginationQuery } from "#/__generated__/core/TreatmentPlanMeasureListPaginationQuery.graphql";
import { LinkedMeasureDialog } from "#/components/measures/LinkedMeasuresDialog";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import MeasureFormDialog from "#/pages/organizations/measures/dialog/MeasureFormDialog";

const PAGE_SIZE = 50;

export const treatmentPlanMeasureListFragment = graphql`
  fragment TreatmentPlanMeasureList_treatmentPlan on TreatmentPlan
  @refetchable(queryName: "TreatmentPlanMeasureListPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    id
    treatment
    canUpdate: permission(action: "core:treatment-plan:update")
    organization {
      canCreateMeasure: permission(action: "core:measure:create")
    }
    measures(first: $first, after: $after)
      @connection(key: "TreatmentPlanMeasureList_measures", filters: []) {
      __id
      edges {
        node {
          id
          ...TreatmentPlanMeasureList_measure
        }
      }
    }
  }
`;

const measureFragment = graphql`
  fragment TreatmentPlanMeasureList_measure on Measure {
    id
    name
    state
  }
`;

const attachMeasureMutation = graphql`
  mutation TreatmentPlanMeasureListCreateMutation(
    $input: CreateTreatmentPlanMeasureMappingInput!
    $connections: [ID!]!
  ) {
    createTreatmentPlanMeasureMapping(input: $input) {
      measureEdge @prependEdge(connections: $connections) {
        node {
          id
          ...TreatmentPlanMeasureList_measure
        }
      }
      treatmentPlanEdge {
        node {
          id
          netLikelihood
          netImpact
          netRiskScore
          measureCount: measures(first: 0) {
            totalCount
          }
          implementedMeasures: measures(
            first: 0
            filter: { state: IMPLEMENTED }
          ) {
            totalCount
          }
          inProgressMeasures: measures(
            first: 0
            filter: { state: IN_PROGRESS }
          ) {
            totalCount
          }
          notImplementedMeasures: measures(
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

const detachMeasureMutation = graphql`
  mutation TreatmentPlanMeasureListDetachMutation(
    $input: DeleteTreatmentPlanMeasureMappingInput!
    $connections: [ID!]!
  ) {
    deleteTreatmentPlanMeasureMapping(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
      deletedTreatmentPlanId
      treatmentPlan {
        id
        netLikelihood
        netImpact
        netRiskScore
        measureCount: measures(first: 0) {
          totalCount
        }
        implementedMeasures: measures(first: 0, filter: { state: IMPLEMENTED }) {
          totalCount
        }
        inProgressMeasures: measures(first: 0, filter: { state: IN_PROGRESS }) {
          totalCount
        }
        notImplementedMeasures: measures(
          first: 0
          filter: { state: NOT_IMPLEMENTED }
        ) {
          totalCount
        }
      }
    }
  }
`;

interface TreatmentPlanMeasureListProps {
  treatmentPlanKey: TreatmentPlanMeasureList_treatmentPlan$key;
  onChanged?: () => void;
}

export function TreatmentPlanMeasureList({
  treatmentPlanKey,
  onChanged,
}: TreatmentPlanMeasureListProps) {
  const { t } = useTranslation();
  const { data: treatmentPlan, hasNext, isLoadingNext, loadNext }
    = usePaginationFragment<
      TreatmentPlanMeasureListPaginationQuery,
      TreatmentPlanMeasureList_treatmentPlan$key
    >(treatmentPlanMeasureListFragment, treatmentPlanKey);
  const measures = treatmentPlan.measures?.edges?.map(edge => edge.node) ?? [];
  const connectionId = treatmentPlan.measures?.__id ?? "";
  const [attachMeasure, isAttaching] = useMutation<TreatmentPlanMeasureListCreateMutation>(
    attachMeasureMutation,
  );
  const [detachMeasure, isDetaching] = useMutation<TreatmentPlanMeasureListDetachMutation>(
    detachMeasureMutation,
  );
  const isLoading = isAttaching || isDetaching;
  const inFlightRef = useRef(false);
  const accepted = treatmentPlan.treatment === "ACCEPTED";
  const readOnly = !treatmentPlan.canUpdate || accepted;

  const onAttach = async (measureId: string) => {
    if (inFlightRef.current || isLoading) {
      return;
    }

    inFlightRef.current = true;
    try {
      await attachMeasure({
        variables: {
          input: {
            measureId,
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

  const onDetach = async (measureId: string) => {
    if (inFlightRef.current || isLoading) {
      return;
    }

    inFlightRef.current = true;
    try {
      await detachMeasure({
        variables: {
          input: {
            measureId,
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
          {t("treatmentPlanMeasureList.title")}
        </h3>
        {!readOnly && (
          <div className="flex items-center gap-3">
            {treatmentPlan.organization.canCreateMeasure && (
              <MeasureFormDialog onCreated={onAttach}>
                <button
                  type="button"
                  disabled={isLoading}
                  className="flex cursor-pointer items-center gap-1 text-sm text-txt-secondary hover:text-txt-primary disabled:opacity-60"
                >
                  <IconPlusLarge size={16} />
                  {t("treatmentPlanMeasureList.actions.create")}
                </button>
              </MeasureFormDialog>
            )}
            <LinkedMeasureDialog
              connectionId={connectionId}
              disabled={isLoading}
              linkedMeasures={measures}
              onLink={(measureId) => {
                void onAttach(measureId);
              }}
              onUnlink={(measureId) => {
                void onDetach(measureId);
              }}
            >
              <button
                type="button"
                disabled={isLoading}
                className="flex cursor-pointer items-center gap-1 text-sm text-txt-secondary hover:text-txt-primary disabled:opacity-60"
              >
                <IconPlusLarge size={16} />
                {t("treatmentPlanMeasureList.actions.link")}
              </button>
            </LinkedMeasureDialog>
          </div>
        )}
      </div>
      {measures.length === 0 && (
        <p className="text-sm text-txt-secondary">
          {accepted
            ? t("treatmentPlanMeasureList.acceptedEmpty")
            : t("treatmentPlanMeasureList.empty")}
        </p>
      )}
      {measures.length > 0 && (
        <ul
          className={
            readOnly
              ? "grid w-full grid-cols-[minmax(0,1fr)_auto] items-center gap-x-8 divide-y divide-border-low"
              : "grid w-full grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-x-8 divide-y divide-border-low"
          }
        >
          {measures.map(measure => (
            <MeasureRow
              key={measure.id}
              measureKey={measure}
              readOnly={readOnly}
              disabled={isLoading}
              onDetach={(measureId) => {
                void onDetach(measureId);
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
          {t("treatmentPlanMeasureList.actions.showMore")}
        </Button>
      )}
    </div>
  );
}

function MeasureRow({
  measureKey,
  readOnly,
  disabled,
  onDetach,
}: {
  measureKey: TreatmentPlanMeasureList_measure$key & { id: string };
  readOnly: boolean;
  disabled: boolean;
  onDetach: (measureId: string) => void;
}) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const measure = useFragment(measureFragment, measureKey);

  return (
    <li className="col-span-full grid grid-cols-subgrid items-center py-2.5">
      <Link
        to={`/organizations/${organizationId}/governance/measures/${measure.id}`}
        className="min-w-0 truncate text-sm text-txt-primary hover:underline"
      >
        {measure.name}
      </Link>
      <MeasureBadge state={measure.state} />
      {!readOnly && (
        <button
          type="button"
          disabled={disabled}
          className="cursor-pointer justify-self-end p-0.5 text-txt-tertiary hover:text-txt-primary disabled:pointer-events-none disabled:opacity-60"
          aria-label={t("treatmentPlanMeasureList.actions.unlink")}
          onClick={() => onDetach(measure.id)}
        >
          <IconTrashCan size={16} />
        </button>
      )}
    </li>
  );
}
