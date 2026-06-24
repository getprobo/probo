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

import { useTranslate } from "@probo/i18n";
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
import { graphql, useMutation } from "react-relay";

import type { CreateThreatDialogMutation } from "#/__generated__/core/CreateThreatDialogMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";

const createThreatMutation = graphql`
  mutation CreateThreatDialogMutation(
    $input: CreateRiskAssessmentThreatInput!
    $connections: [ID!]!
  ) {
    createRiskAssessmentThreat(input: $input) {
      riskAssessmentThreatEdge @appendEdge(connections: $connections) {
        node { id processId name category }
      }
    }
  }
`;

export function CreateThreatDialog(props: {
  scopeId: string;
  processes: { id: string; name: string }[];
  connectionId: string;
}) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [createThreat, isCreating] = useMutation<CreateThreatDialogMutation>(createThreatMutation);
  const { register, control, handleSubmit, reset, formState } = useForm({
    defaultValues: { name: "", processId: "", category: "Confidentiality" },
  });
  const onSubmit = (data: { name: string; processId: string; category: string }) => {
    createThreat({
      variables: {
        input: {
          riskAssessmentScopeId: props.scopeId,
          processId: data.processId,
          name: data.name,
          category: data.category,
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
      trigger={<Button icon={IconPlusLarge} variant="secondary" disabled={props.processes.length === 0}>{__("Add")}</Button>}
      title={<Breadcrumb items={[__("Threats"), __("Add Threat")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <ControlledField label={__("Process")} name="processId" control={control} rules={{ required: __("This field is required") }} type="select" placeholder={__("Select process")}>
            {props.processes.map(p => <Option key={p.id} value={p.id}>{p.name}</Option>)}
          </ControlledField>
          <Field label={__("Name")} {...register("name", { required: __("This field is required") })} type="text" error={formState.errors.name?.message} placeholder={__("e.g. SQL injection")} />
          <Field label={__("Category")} {...register("category", { required: __("This field is required") })} type="text" error={formState.errors.category?.message} placeholder={__("e.g. Confidentiality")} />
        </DialogContent>
        <DialogFooter><Button type="submit" disabled={isCreating || props.processes.length === 0}>{__("Add")}</Button></DialogFooter>
      </form>
    </Dialog>
  );
}
