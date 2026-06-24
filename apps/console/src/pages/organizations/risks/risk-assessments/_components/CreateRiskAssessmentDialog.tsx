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
