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

import type { CreateNodeDialogMutation } from "#/__generated__/core/CreateNodeDialogMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";

const createNodeMutation = graphql`
  mutation CreateNodeDialogMutation(
    $input: CreateRiskAssessmentNodeInput!
    $connections: [ID!]!
  ) {
    createRiskAssessmentNode(input: $input) {
      riskAssessmentNodeEdge @appendEdge(connections: $connections) {
        node { id nodeType name }
      }
    }
  }
`;

export function CreateNodeDialog(props: { scopeId: string; connectionId: string }) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [createNode, isCreating] = useMutation<CreateNodeDialogMutation>(createNodeMutation);
  const { register, control, handleSubmit, reset, formState } = useForm({
    defaultValues: { name: "", nodeType: "ASSET" },
  });
  const onSubmit = (data: { name: string; nodeType: string }) => {
    createNode({
      variables: {
        input: {
          riskAssessmentScopeId: props.scopeId,
          nodeType: data.nodeType as "ENTITY" | "BOUNDARY" | "ASSET" | "DATA",
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
      trigger={<Button icon={IconPlusLarge} variant="secondary">{__("Add")}</Button>}
      title={<Breadcrumb items={[__("Nodes"), __("Add Node")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <ControlledField label={__("Type")} name="nodeType" control={control} type="select">
            <Option value="ENTITY">{__("Entity")}</Option>
            <Option value="BOUNDARY">{__("Boundary")}</Option>
            <Option value="ASSET">{__("Asset")}</Option>
            <Option value="DATA">{__("Data")}</Option>
          </ControlledField>
          <Field label={__("Name")} {...register("name", { required: __("This field is required") })} type="text" error={formState.errors.name?.message} />
        </DialogContent>
        <DialogFooter><Button type="submit" disabled={isCreating}>{__("Add")}</Button></DialogFooter>
      </form>
    </Dialog>
  );
}
