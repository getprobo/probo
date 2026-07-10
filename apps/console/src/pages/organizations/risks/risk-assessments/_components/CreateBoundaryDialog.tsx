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
  const { __ } = useTranslate();
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
      trigger={<Button icon={IconPlusLarge} variant="secondary">{__("Add")}</Button>}
      title={<Breadcrumb items={[__("Boundaries"), __("Add Boundary")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field label={__("Name")} {...register("name", { required: __("This field is required") })} type="text" error={formState.errors.name?.message} />
          <ControlledField label={__("Parent boundary")} name="parentBoundaryId" control={control} type="select">
            <Option value="none">{__("None (top level)")}</Option>
            {props.boundaries.map(b => (
              <Option key={b.id} value={b.id}>{b.name}</Option>
            ))}
          </ControlledField>
        </DialogContent>
        <DialogFooter><Button type="submit" disabled={isCreating}>{__("Add")}</Button></DialogFooter>
      </form>
    </Dialog>
  );
}
