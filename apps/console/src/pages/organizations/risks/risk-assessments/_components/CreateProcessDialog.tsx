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
  DialogContent,
  DialogFooter,
  Field,
  IconPlusLarge,
  Option,
  useDialogRef,
} from "@probo/ui";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useMutation } from "react-relay";

import type { CreateProcessDialogMutation } from "#/__generated__/core/CreateProcessDialogMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";

const createProcessMutation = graphql`
  mutation CreateProcessDialogMutation(
    $input: CreateRiskAssessmentProcessInput!
    $connections: [ID!]!
  ) {
    createRiskAssessmentProcess(input: $input) {
      riskAssessmentProcessEdge @appendEdge(connections: $connections) {
        node { id sourceNodeId targetNodeId name }
      }
    }
  }
`;

export function CreateProcessDialog(props: {
  scopeId: string;
  nodes: { id: string; name: string }[];
  connectionId: string;
}) {
  const { t } = useTranslation();
  const dialogRef = useDialogRef();
  const [createProcess, isCreating] = useMutation<CreateProcessDialogMutation>(createProcessMutation);
  const { register, control, handleSubmit, reset, formState } = useForm({
    defaultValues: { name: "", sourceNodeId: "", targetNodeId: "" },
  });
  const onSubmit = (data: { name: string; sourceNodeId: string; targetNodeId: string }) => {
    createProcess({
      variables: {
        input: {
          riskAssessmentScopeId: props.scopeId,
          sourceNodeId: data.sourceNodeId,
          targetNodeId: data.targetNodeId,
          name: data.name,
        },
        connections: [props.connectionId],
      },
      onCompleted: () => {
        reset();
        dialogRef.current?.close();
      },
    });
  };
  return (
    <Dialog
      className="max-w-lg"
      ref={dialogRef}
      trigger={<Button icon={IconPlusLarge} variant="secondary" disabled={props.nodes.length < 2}>{t("createRiskAssessmentProcessDialog.actions.add")}</Button>}
      title={<Breadcrumb items={[t("createRiskAssessmentProcessDialog.breadcrumb.processes"), t("createRiskAssessmentProcessDialog.breadcrumb.addProcess")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <ControlledField label={t("createRiskAssessmentProcessDialog.fields.source")} name="sourceNodeId" control={control} rules={{ required: t("createRiskAssessmentProcessDialog.validation.required") }} type="select" placeholder={t("createRiskAssessmentProcessDialog.placeholders.sourceNode")}>
            {props.nodes.map(n => <Option key={n.id} value={n.id}>{n.name}</Option>)}
          </ControlledField>
          <ControlledField label={t("createRiskAssessmentProcessDialog.fields.target")} name="targetNodeId" control={control} rules={{ required: t("createRiskAssessmentProcessDialog.validation.required") }} type="select" placeholder={t("createRiskAssessmentProcessDialog.placeholders.targetNode")}>
            {props.nodes.map(n => <Option key={n.id} value={n.id}>{n.name}</Option>)}
          </ControlledField>
          <Field label={t("createRiskAssessmentProcessDialog.fields.name")} {...register("name", { required: t("createRiskAssessmentProcessDialog.validation.required") })} type="text" error={formState.errors.name?.message} placeholder={t("createRiskAssessmentProcessDialog.placeholders.name")} />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isCreating || props.nodes.length < 2}>{t("createRiskAssessmentProcessDialog.actions.add")}</Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
