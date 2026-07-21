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

import type { CreateBoundaryDialogMutation } from "#/__generated__/core/CreateBoundaryDialogMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";

const createBoundaryMutation = graphql`
  mutation CreateBoundaryDialogMutation(
    $input: CreateRiskAssessmentBoundaryInput!
    $connections: [ID!]!
  ) {
    createRiskAssessmentBoundary(input: $input) {
      riskAssessmentBoundaryEdge @appendEdge(connections: $connections) {
        node { id name parentBoundaryId }
      }
    }
  }
`;

export function CreateBoundaryDialog(props: {
  scopeId: string;
  connectionId: string;
  boundaries: { id: string; name: string }[];
}) {
  const { t } = useTranslation();
  const dialogRef = useDialogRef();
  const [createBoundary, isCreating] = useMutation<CreateBoundaryDialogMutation>(createBoundaryMutation);
  const { register, control, handleSubmit, reset, formState } = useForm({
    defaultValues: { name: "", parentBoundaryId: "none" },
  });
  const onSubmit = (data: { name: string; parentBoundaryId: string }) => {
    createBoundary({
      variables: {
        input: {
          riskAssessmentScopeId: props.scopeId,
          name: data.name,
          parentBoundaryId: data.parentBoundaryId === "none" ? null : data.parentBoundaryId,
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
      trigger={<Button icon={IconPlusLarge} variant="secondary">{t("createRiskAssessmentBoundaryDialog.actions.add")}</Button>}
      title={<Breadcrumb items={[t("createRiskAssessmentBoundaryDialog.breadcrumb.boundaries"), t("createRiskAssessmentBoundaryDialog.breadcrumb.addBoundary")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field label={t("createRiskAssessmentBoundaryDialog.fields.name")} {...register("name", { required: t("createRiskAssessmentBoundaryDialog.validation.nameRequired") })} type="text" error={formState.errors.name?.message} />
          <ControlledField label={t("createRiskAssessmentBoundaryDialog.fields.parentBoundary")} name="parentBoundaryId" control={control} type="select">
            <Option value="none">{t("createRiskAssessmentBoundaryDialog.noneTopLevel")}</Option>
            {props.boundaries.map(b => (
              <Option key={b.id} value={b.id}>{b.name}</Option>
            ))}
          </ControlledField>
        </DialogContent>
        <DialogFooter><Button type="submit" disabled={isCreating}>{t("createRiskAssessmentBoundaryDialog.actions.add")}</Button></DialogFooter>
      </form>
    </Dialog>
  );
}
