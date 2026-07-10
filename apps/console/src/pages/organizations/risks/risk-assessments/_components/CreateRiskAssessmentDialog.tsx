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

import type { CreateRiskAssessmentDialogCreateMutation } from "#/__generated__/core/CreateRiskAssessmentDialogCreateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const createMutation = graphql`
  mutation CreateRiskAssessmentDialogCreateMutation(
    $input: CreateRiskAssessmentInput!
    $connections: [ID!]!
  ) {
    createRiskAssessment(input: $input) {
      riskAssessmentEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          createdAt
        }
      }
    }
  }
`;

export function CreateRiskAssessmentDialog(props: {
  connectionId: string;
}) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();
  const [createRiskAssessment, isCreating] = useMutation<CreateRiskAssessmentDialogCreateMutation>(createMutation);
  const { register, handleSubmit, reset, formState } = useForm({
    defaultValues: { name: "", description: "" },
  });

  const onSubmit = (data: { name: string; description: string }) => {
    createRiskAssessment({
      variables: {
        input: {
          organizationId,
          name: data.name,
          description: data.description || null,
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
      trigger={(
        <Button icon={IconPlusLarge} variant="primary">
          {__("New Risk Assessment")}
        </Button>
      )}
      title={(
        <Breadcrumb
          items={[__("Risk Assessments"), __("New Risk Assessment")]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Name")}
            {...register("name", { required: __("This field is required") })}
            type="text"
            error={formState.errors.name?.message}
            placeholder={__("e.g. Platform Threat Model 2026")}
          />
          <Field
            label={__("Description")}
            {...register("description")}
            type="textarea"
            rows={3}
            placeholder={__("Describe the scope and purpose of this assessment...")}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isCreating}>
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
