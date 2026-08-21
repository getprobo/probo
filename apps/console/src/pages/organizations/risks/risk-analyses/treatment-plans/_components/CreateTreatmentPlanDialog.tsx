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

import {
  Breadcrumb,
  Button,
  Dialog,
  DialogFooter,
  Field,
  IconPlusLarge,
  useDialogRef,
} from "@probo/ui";
import { type ReactNode, Suspense, useCallback, useEffect, useRef, useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type {
  CreateTreatmentPlanDialogCreateMutation,
  RiskTreatment,
} from "#/__generated__/core/CreateTreatmentPlanDialogCreateMutation.graphql";
import type { PrefillFromRiskQuery } from "#/__generated__/core/PrefillFromRiskQuery.graphql";
import type { ScenarioLinkedRiskSelectQuery } from "#/__generated__/core/ScenarioLinkedRiskSelectQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import type { MatrixSize } from "../../_components/matrixSize";

import { PrefillFromRisk, prefillFromRiskQuery, type TreatmentPlanPrefillValues } from "./PrefillFromRisk";
import { ScenarioLinkedRiskSelect, scenarioLinkedRiskSelectQuery } from "./ScenarioLinkedRiskSelect";
import { TreatmentPlanDialogFields } from "./TreatmentPlanDialogFields";

const createMutation = graphql`
  mutation CreateTreatmentPlanDialogCreateMutation(
    $input: CreateTreatmentPlanInput!
    $connections: [ID!]!
  ) {
    createTreatmentPlan(input: $input) {
      treatmentPlanEdge @prependEdge(connections: $connections) {
        node {
          id
          inherentLikelihood
          inherentImpact
          netLikelihood
          netImpact
          residualLikelihood
          residualImpact
          risk {
            id
            name
          }
          ...TreatmentPlanListItem_treatmentPlan
        }
      }
    }
  }
`;

type FormData = {
  riskId: string;
  treatment: RiskTreatment;
  ownerId: string;
  inherentLikelihood: string;
  inherentImpact: string;
  residualLikelihood: string;
  residualImpact: string;
};

interface CreateTreatmentPlanDialogProps {
  connectionId: string;
  matrixSize: MatrixSize;
  riskId?: string;
  children?: ReactNode;
  onCompleted?: () => void;
}

export function CreateTreatmentPlanDialog({
  connectionId,
  matrixSize,
  riskId: lockedRiskId,
  children,
  onCompleted,
}: CreateTreatmentPlanDialogProps) {
  const { t } = useTranslation();
  const { riskAnalysisId } = useParams<{ riskAnalysisId: string }>();
  const dialogRef = useDialogRef();
  const [open, setOpen] = useState(false);
  const [createTreatmentPlan, isCreating]
    = useMutation<CreateTreatmentPlanDialogCreateMutation>(createMutation);
  const [eligibleQueryRef, loadEligibleQuery]
    = useQueryLoader<ScenarioLinkedRiskSelectQuery>(scenarioLinkedRiskSelectQuery);
  const [riskQueryRef, loadRiskQuery]
    = useQueryLoader<PrefillFromRiskQuery>(prefillFromRiskQuery);
  const { control, handleSubmit, reset, resetField, setValue, setError, formState } = useForm<FormData>({
    defaultValues: {
      riskId: lockedRiskId ?? "",
      treatment: "MITIGATED",
      ownerId: "",
      inherentLikelihood: "1",
      inherentImpact: "1",
      residualLikelihood: "1",
      residualImpact: "1",
    },
  });
  const riskId = useWatch({ control, name: "riskId" });
  const inherentLikelihood = useWatch({ control, name: "inherentLikelihood" });
  const inherentImpact = useWatch({ control, name: "inherentImpact" });
  const residualDirty = Boolean(
    formState.dirtyFields.residualLikelihood || formState.dirtyFields.residualImpact,
  );
  const dirtyFieldsRef = useRef(formState.dirtyFields);
  const [prefilledRiskId, setPrefilledRiskId] = useState("");
  const prefillPending = Boolean(riskId) && prefilledRiskId !== riskId;
  const matchingRiskQueryRef = riskQueryRef?.variables.riskId === riskId
    ? riskQueryRef
    : null;

  useEffect(() => {
    dirtyFieldsRef.current = formState.dirtyFields;
  });

  useEffect(() => {
    resetField("treatment");
    resetField("ownerId");
    resetField("inherentLikelihood");
    resetField("inherentImpact");
    resetField("residualLikelihood");
    resetField("residualImpact");
  }, [resetField, riskId]);

  useEffect(() => {
    if (open && riskId) {
      loadRiskQuery({ riskId });
    }
  }, [loadRiskQuery, open, riskId]);

  useEffect(() => {
    if (residualDirty) {
      return;
    }

    setValue("residualLikelihood", inherentLikelihood);
    setValue("residualImpact", inherentImpact);
  }, [inherentImpact, inherentLikelihood, residualDirty, setValue]);

  const onPrefill = useCallback(
    (values: TreatmentPlanPrefillValues) => {
      if (values.riskId !== riskId) {
        return;
      }

      const dirty = dirtyFieldsRef.current;
      if (!dirty.treatment) {
        setValue("treatment", values.treatment);
      }
      if (!dirty.ownerId) {
        setValue("ownerId", values.ownerId);
      }
      setPrefilledRiskId(values.riskId);
    },
    [riskId, setValue],
  );

  const onSubmit = async (data: FormData) => {
    if (!data.riskId) {
      setError("riskId", {
        type: "required",
        message: t("createTreatmentPlanDialog.validation.riskRequired"),
      });
      return;
    }

    if (!data.ownerId) {
      setError("ownerId", {
        type: "required",
        message: t("createTreatmentPlanDialog.validation.ownerRequired"),
      });
      return;
    }

    if (!riskAnalysisId) {
      return;
    }

    try {
      await createTreatmentPlan({
        variables: {
          input: {
            riskId: data.riskId,
            riskAnalysisId,
            treatment: data.treatment,
            ownerId: data.ownerId,
            inherentLikelihood: Number(data.inherentLikelihood),
            inherentImpact: Number(data.inherentImpact),
            residualLikelihood: Number(data.residualLikelihood),
            residualImpact: Number(data.residualImpact),
          },
          connections: [connectionId],
        },
      });
      reset({
        riskId: lockedRiskId ?? "",
        treatment: "MITIGATED",
        ownerId: "",
        inherentLikelihood: "1",
        inherentImpact: "1",
        residualLikelihood: "1",
        residualImpact: "1",
      });
      dialogRef.current?.close();
      onCompleted?.();
    } catch {
      // Error toast is handled by useMutation.
    }
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children ?? (
        <Button icon={IconPlusLarge} variant="secondary">
          {t("createTreatmentPlanDialog.actions.add")}
        </Button>
      )}
      title={(
        <Breadcrumb
          items={[
            t("createTreatmentPlanDialog.breadcrumb.treatmentPlans"),
            t("createTreatmentPlanDialog.breadcrumb.new"),
          ]}
        />
      )}
      onOpenChange={(nextOpen) => {
        setOpen(nextOpen);
        if (!nextOpen) {
          return;
        }
        if (riskAnalysisId && !lockedRiskId) {
          loadEligibleQuery({ riskAnalysisId });
        }
      }}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <TreatmentPlanDialogFields
          control={control}
          copy="createTreatmentPlanDialog"
          disabled={prefillPending}
          matrixSize={matrixSize}
        >
          {lockedRiskId
            ? null
            : (
                <Suspense
                  fallback={(
                    <Field
                      disabled
                      label={t("createTreatmentPlanDialog.fields.risk")}
                      type="select"
                    />
                  )}
                >
                  {eligibleQueryRef
                    ? (
                        <ScenarioLinkedRiskSelect
                          control={control}
                          error={formState.errors.riskId?.message}
                          queryRef={eligibleQueryRef}
                        />
                      )
                    : null}
                </Suspense>
              )}
          {matchingRiskQueryRef
            ? (
                <Suspense fallback={null}>
                  <PrefillFromRisk
                    key={riskId}
                    queryRef={matchingRiskQueryRef}
                    riskId={riskId}
                    onPrefill={onPrefill}
                  />
                </Suspense>
              )
            : null}
        </TreatmentPlanDialogFields>
        <DialogFooter>
          <Button type="submit" disabled={isCreating || prefillPending}>
            {t("createTreatmentPlanDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
