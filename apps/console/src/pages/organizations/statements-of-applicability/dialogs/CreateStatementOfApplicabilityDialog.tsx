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

import { formatError, type GraphQLError } from "@probo/helpers";
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
            error as GraphQLError,
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
