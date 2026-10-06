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

import type { TreatmentPlanInternalControlsDialogQuery } from "#/__generated__/core/TreatmentPlanInternalControlsDialogQuery.graphql";

import { TreatmentPlanInternalControlList } from "./TreatmentPlanInternalControlList";

const treatmentPlanInternalControlsDialogQuery = graphql`
  query TreatmentPlanInternalControlsDialogQuery($treatmentPlanId: ID!) {
    node(id: $treatmentPlanId) @required(action: THROW) {
      __typename
      ... on TreatmentPlan {
        netLikelihood
        netImpact
        netRiskScore
        ...TreatmentPlanInternalControlList_meta
        ...TreatmentPlanInternalControlList_treatmentPlan
      }
    }
  }
`;

interface TreatmentPlanInternalControlsDialogProps {
  dialogRef: DialogRef;
  treatmentPlanId: string;
}

export function TreatmentPlanInternalControlsDialog({
  dialogRef,
  treatmentPlanId,
}: TreatmentPlanInternalControlsDialogProps) {
  const { t } = useTranslation();
  const [queryRef, loadQuery] = useQueryLoader<TreatmentPlanInternalControlsDialogQuery>(
    treatmentPlanInternalControlsDialogQuery,
  );

  const reload = useCallback((policy: "store-and-network" | "network-only") => {
    loadQuery({ treatmentPlanId }, { fetchPolicy: policy });
  }, [loadQuery, treatmentPlanId]);

  return (
    <Dialog
      ref={dialogRef}
      title={t("treatmentPlanInternalControlsDialog.title")}
      onOpenChange={(open) => {
        if (open) {
          reload("store-and-network");
        }
      }}
    >
      {queryRef
        ? (
            <Suspense fallback={<p className="p-6 text-sm text-txt-secondary">{t("treatmentPlanInternalControlsDialog.loading")}</p>}>
              <TreatmentPlanInternalControlsDialogBody
                queryRef={queryRef}
                onCompleted={() => reload("network-only")}
              />
            </Suspense>
          )
        : null}
    </Dialog>
  );
}

function TreatmentPlanInternalControlsDialogBody({
  queryRef,
  onCompleted,
}: {
  queryRef: PreloadedQuery<TreatmentPlanInternalControlsDialogQuery>;
  onCompleted: () => void;
}) {
  const { t } = useTranslation();
  const data = usePreloadedQuery<TreatmentPlanInternalControlsDialogQuery>(
    treatmentPlanInternalControlsDialogQuery,
    queryRef,
  );
  if (data.node.__typename !== "TreatmentPlan") {
    throw new Error("invalid node type");
  }

  return (
    <DialogContent className="p-4">
      <p className="mb-3 text-sm text-txt-secondary">
        {t("treatmentPlanInternalControlsDialog.netScore", {
          score: data.node.netRiskScore,
          likelihood: data.node.netLikelihood,
          impact: data.node.netImpact,
        })}
      </p>
      <TreatmentPlanInternalControlList
        treatmentPlanKey={data.node}
        onChanged={onCompleted}
      />
    </DialogContent>
  );
}
