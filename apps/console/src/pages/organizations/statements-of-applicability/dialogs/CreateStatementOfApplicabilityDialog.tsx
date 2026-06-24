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

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
  useToast,
} from "@probo/ui";
import type { ReactNode } from "react";
import { useMutation } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CreateStatementOfApplicabilityDialogMutation } from "#/__generated__/core/CreateStatementOfApplicabilityDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const createMutation = graphql`
    mutation CreateStatementOfApplicabilityDialogMutation(
        $input: CreateStatementOfApplicabilityInput!
        $connections: [ID!]!
    ) {
        createStatementOfApplicability(input: $input) {
            statementOfApplicabilityEdge @prependEdge(connections: $connections) {
                node {
                    id
                    name
                    createdAt
                    updatedAt
                    canDelete: permission(action: "core:statement-of-applicability:delete")
                    ...StatementOfApplicabilityRowFragment
                }
            }
        }
    }
`;

type Props = {
  children: ReactNode;
  connectionId: string;
};

const schema = z.object({
  name: z.string().min(1),
});

export function CreateStatementOfApplicabilityDialog({
  children,
  connectionId,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { register, handleSubmit, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        name: "",
      },
    },
  );
  const ref = useDialogRef();

  const [createStatementOfApplicability, isCreating]
    = useMutation<CreateStatementOfApplicabilityDialogMutation>(createMutation);

  const onSubmit = (data: z.infer<typeof schema>) => {
    createStatementOfApplicability({
      variables: {
        input: {
          name: data.name,
          organizationId,
        },
        connections: [connectionId],
      },
      onCompleted(response) {
        toast({
          title: __("Success"),
          description: __("Statement of applicability created successfully."),
          variant: "success",
        });
        reset();
        ref.current?.close();
        const statementOfApplicabilityId
          = response.createStatementOfApplicability.statementOfApplicabilityEdge
            .node.id;
        void navigate(
          `/organizations/${organizationId}/statements-of-applicability/${statementOfApplicabilityId}`,
        );
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to create statement of applicability"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={(
        <Breadcrumb
          items={[
            __("Statements of Applicability"),
            __("New Statement of Applicability"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Name")}
            {...register("name")}
            type="text"
            required
          />
        </DialogContent>
        <DialogFooter>
          <Button disabled={isCreating} type="submit">
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
