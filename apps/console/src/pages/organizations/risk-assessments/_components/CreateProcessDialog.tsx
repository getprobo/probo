// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
  const { __ } = useTranslate();
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
      trigger={<Button icon={IconPlusLarge} variant="secondary" disabled={props.nodes.length < 2}>{__("Add")}</Button>}
      title={<Breadcrumb items={[__("Processes"), __("Add Process")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <ControlledField label={__("Source")} name="sourceNodeId" control={control} rules={{ required: __("This field is required") }} type="select" placeholder={__("Select source node")}>
            {props.nodes.map(n => <Option key={n.id} value={n.id}>{n.name}</Option>)}
          </ControlledField>
          <ControlledField label={__("Target")} name="targetNodeId" control={control} rules={{ required: __("This field is required") }} type="select" placeholder={__("Select target node")}>
            {props.nodes.map(n => <Option key={n.id} value={n.id}>{n.name}</Option>)}
          </ControlledField>
          <Field label={__("Name")} {...register("name", { required: __("This field is required") })} type="text" error={formState.errors.name?.message} placeholder={__("e.g. User → API")} />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isCreating || props.nodes.length < 2}>{__("Add")}</Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
