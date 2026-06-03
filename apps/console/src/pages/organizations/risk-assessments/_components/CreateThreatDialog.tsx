// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
