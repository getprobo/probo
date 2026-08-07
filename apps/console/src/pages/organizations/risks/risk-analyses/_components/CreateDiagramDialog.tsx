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
  useDialogRef,
} from "@probo/ui";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useMutation } from "react-relay";
import { useParams } from "react-router";

import type { CreateDiagramDialogMutation } from "#/__generated__/core/CreateDiagramDialogMutation.graphql";

const createDiagramMutation = graphql`
  mutation CreateDiagramDialogMutation(
    $input: CreateRiskAnalysisDiagramInput!
    $connections: [ID!]!
  ) {
    createRiskAnalysisDiagram(input: $input) {
      riskAnalysisDiagramEdge @appendEdge(connections: $connections) {
        node {
          id
          ...DiagramCardFragment
        }
      }
    }
  }
`;

export function CreateDiagramDialog(props: { connectionId: string }) {
  const { riskAnalysisId } = useParams<{ riskAnalysisId: string }>();
  const { t } = useTranslation();
  const dialogRef = useDialogRef();
  const [createDiagram, isCreating] = useMutation<CreateDiagramDialogMutation>(createDiagramMutation);
  const { register, handleSubmit, reset, formState } = useForm({
    defaultValues: { name: "" },
  });
  const onSubmit = (data: { name: string }) => {
    if (!riskAnalysisId) return;
    createDiagram({
      variables: {
        input: { riskAnalysisId, name: data.name },
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
      trigger={<Button icon={IconPlusLarge} variant="secondary">{t("createRiskAnalysisDiagramDialog.actions.addDiagram")}</Button>}
      title={<Breadcrumb items={[t("createRiskAnalysisDiagramDialog.breadcrumb.diagrams"), t("createRiskAnalysisDiagramDialog.breadcrumb.newDiagram")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field label={t("createRiskAnalysisDiagramDialog.fields.name")} {...register("name", { required: t("createRiskAnalysisDiagramDialog.validation.nameRequired") })} type="text" error={formState.errors.name?.message} placeholder={t("createRiskAnalysisDiagramDialog.placeholders.name")} />
        </DialogContent>
        <DialogFooter><Button type="submit" disabled={isCreating}>{t("createRiskAnalysisDiagramDialog.actions.create")}</Button></DialogFooter>
      </form>
    </Dialog>
  );
}
