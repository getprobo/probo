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

import type { CreateNodeDialogMutation } from "#/__generated__/core/CreateNodeDialogMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";

const createNodeMutation = graphql`
  mutation CreateNodeDialogMutation(
    $input: CreateRiskAssessmentNodeInput!
    $connections: [ID!]!
  ) {
    createRiskAssessmentNode(input: $input) {
      riskAssessmentNodeEdge @appendEdge(connections: $connections) {
        node { id nodeType name boundaryId }
      }
    }
  }
`;

export function CreateNodeDialog(props: {
  scopeId: string;
  connectionId: string;
  boundaries: { id: string; name: string }[];
}) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [createNode, isCreating] = useMutation<CreateNodeDialogMutation>(createNodeMutation);
  const { register, control, handleSubmit, reset, formState } = useForm({
    defaultValues: { name: "", nodeType: "ASSET", boundaryId: "none" },
  });
  const onSubmit = (data: { name: string; nodeType: string; boundaryId: string }) => {
    createNode({
      variables: {
        input: {
          riskAssessmentScopeId: props.scopeId,
          nodeType: data.nodeType as "ENTITY" | "ASSET" | "DATA",
          name: data.name,
          boundaryId: data.boundaryId === "none" ? null : data.boundaryId,
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
            <Option value="ASSET">{__("Asset")}</Option>
            <Option value="DATA">{__("Data")}</Option>
          </ControlledField>
          <Field label={__("Name")} {...register("name", { required: __("This field is required") })} type="text" error={formState.errors.name?.message} />
          {props.boundaries.length > 0 && (
            <ControlledField label={__("Boundary")} name="boundaryId" control={control} type="select">
              <Option value="none">{__("None")}</Option>
              {props.boundaries.map(b => (
                <Option key={b.id} value={b.id}>{b.name}</Option>
              ))}
            </ControlledField>
          )}
        </DialogContent>
        <DialogFooter><Button type="submit" disabled={isCreating}>{__("Add")}</Button></DialogFooter>
      </form>
    </Dialog>
  );
}
