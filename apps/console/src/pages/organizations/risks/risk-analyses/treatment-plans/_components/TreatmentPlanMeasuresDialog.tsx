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

import { Dialog, DialogContent, type DialogRef } from "@probo/ui";
import { Suspense, useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { graphql } from "relay-runtime";

import type { TreatmentPlanMeasuresDialogQuery } from "#/__generated__/core/TreatmentPlanMeasuresDialogQuery.graphql";

import { TreatmentPlanMeasureList } from "./TreatmentPlanMeasureList";

const treatmentPlanMeasuresDialogQuery = graphql`
  query TreatmentPlanMeasuresDialogQuery($treatmentPlanId: ID!) {
    node(id: $treatmentPlanId) @required(action: THROW) {
      __typename
      ... on TreatmentPlan {
        netLikelihood
        netImpact
        netRiskScore
        ...TreatmentPlanMeasureList_treatmentPlan
      }
    }
  }
`;

interface TreatmentPlanMeasuresDialogProps {
  dialogRef: DialogRef;
  treatmentPlanId: string;
}

export function TreatmentPlanMeasuresDialog({
  dialogRef,
  treatmentPlanId,
}: TreatmentPlanMeasuresDialogProps) {
  const { t } = useTranslation();
  const [queryRef, loadQuery] = useQueryLoader<TreatmentPlanMeasuresDialogQuery>(
    treatmentPlanMeasuresDialogQuery,
  );

  const reload = useCallback((policy: "store-and-network" | "network-only") => {
    loadQuery({ treatmentPlanId }, { fetchPolicy: policy });
  }, [loadQuery, treatmentPlanId]);

  return (
    <Dialog
      ref={dialogRef}
      title={t("treatmentPlanMeasuresDialog.title")}
      onOpenChange={(open) => {
        if (open) {
          reload("store-and-network");
        }
      }}
    >
      {queryRef
        ? (
            <Suspense fallback={<p className="p-6 text-sm text-txt-secondary">{t("treatmentPlanMeasuresDialog.loading")}</p>}>
              <TreatmentPlanMeasuresDialogBody
                queryRef={queryRef}
                onCompleted={() => reload("network-only")}
              />
            </Suspense>
          )
        : null}
    </Dialog>
  );
}

function TreatmentPlanMeasuresDialogBody({
  queryRef,
  onCompleted,
}: {
  queryRef: PreloadedQuery<TreatmentPlanMeasuresDialogQuery>;
  onCompleted: () => void;
}) {
  const { t } = useTranslation();
  const data = usePreloadedQuery<TreatmentPlanMeasuresDialogQuery>(
    treatmentPlanMeasuresDialogQuery,
    queryRef,
  );
  if (data.node.__typename !== "TreatmentPlan") {
    throw new Error("invalid node type");
  }

  return (
    <DialogContent className="p-4">
      <p className="mb-3 text-sm text-txt-secondary">
        {t("treatmentPlanMeasuresDialog.netScore", {
          score: data.node.netRiskScore,
          likelihood: data.node.netLikelihood,
          impact: data.node.netImpact,
        })}
      </p>
      <TreatmentPlanMeasureList
        treatmentPlanKey={data.node}
        onChanged={onCompleted}
      />
    </DialogContent>
  );
}
