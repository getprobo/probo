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
  useDialogRef,
} from "@probo/ui";
import { useForm } from "react-hook-form";
import { graphql, useMutation } from "react-relay";
import { useParams } from "react-router";

import type { CreateScopeDialogMutation } from "#/__generated__/core/CreateScopeDialogMutation.graphql";

const createScopeMutation = graphql`
  mutation CreateScopeDialogMutation(
    $input: CreateRiskAssessmentScopeInput!
    $connections: [ID!]!
  ) {
    createRiskAssessmentScope(input: $input) {
      riskAssessmentScopeEdge @appendEdge(connections: $connections) {
        node {
          id
          ...ScopeCardFragment
        }
      }
    }
  }
`;

export function CreateScopeDialog(props: { connectionId: string }) {
  const { riskAssessmentId } = useParams<{ riskAssessmentId: string }>();
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [createScope, isCreating] = useMutation<CreateScopeDialogMutation>(createScopeMutation);
  const { register, handleSubmit, reset, formState } = useForm({
    defaultValues: { name: "" },
  });
  const onSubmit = (data: { name: string }) => {
    if (!riskAssessmentId) return;
    createScope({
      variables: {
        input: { riskAssessmentId, name: data.name },
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
      trigger={<Button icon={IconPlusLarge} variant="secondary">{__("Add Scope")}</Button>}
      title={<Breadcrumb items={[__("Scopes"), __("New Scope")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field label={__("Name")} {...register("name", { required: __("This field is required") })} type="text" error={formState.errors.name?.message} placeholder={__("e.g. API layer")} />
        </DialogContent>
        <DialogFooter><Button type="submit" disabled={isCreating}>{__("Create")}</Button></DialogFooter>
      </form>
    </Dialog>
  );
}
