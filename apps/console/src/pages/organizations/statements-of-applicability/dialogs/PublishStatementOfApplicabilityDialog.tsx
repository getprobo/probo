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
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconSend,
  IconUpload,
  useDialogRef,
  useToast,
} from "@probo/ui";
import type { ReactNode } from "react";
import { useMemo } from "react";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { PublishStatementOfApplicabilityDialogMutation } from "#/__generated__/core/PublishStatementOfApplicabilityDialogMutation.graphql";
import { PeopleMultiSelectField } from "#/components/form/PeopleMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const publishMutation = graphql`
  mutation PublishStatementOfApplicabilityDialogMutation(
    $input: PublishStatementOfApplicabilityInput!
  ) {
    publishStatementOfApplicability(input: $input) {
      documentEdge {
        node {
          id
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  statementOfApplicabilityId: string;
  defaultApproverIds?: string[];
};

export function PublishStatementOfApplicabilityDialog({
  children,
  statementOfApplicabilityId,
  defaultApproverIds = [],
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();

  const schema = useMemo(() => z.object({
    approverIds: z.array(z.string()),
  }), []);

  const {
    control,
    handleSubmit,
    reset,
    watch,
  } = useFormWithSchema(schema, {
    defaultValues: {
      approverIds: defaultApproverIds,
    },
  });

  const [publish, isPublishing]
    = useMutation<PublishStatementOfApplicabilityDialogMutation>(publishMutation);

  const approverIds = watch("approverIds");
  const hasApprovers = approverIds.length > 0;

  const onSubmit = (data: z.infer<typeof schema>) => {
    publish({
      variables: {
        input: {
          statementOfApplicabilityId,
          approverIds: data.approverIds.length > 0 ? data.approverIds : undefined,
        },
      },
      onCompleted(response) {
        const documentId = response.publishStatementOfApplicability?.documentEdge?.node?.id;
        if (documentId) {
          toast({
            title: __("Success"),
            description: hasApprovers
              ? __("Approval requested successfully.")
              : __("Statement of Applicability published successfully."),
            variant: "success",
          });
        }
        dialogRef.current?.close();
        reset();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to publish Statement of Applicability"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  return (
    <Dialog
      className="max-w-xl"
      ref={dialogRef}
      trigger={children}
      title={__("Publish Statement of Applicability")}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded>
          <div className="space-y-4">
            <p className="text-sm text-txt-secondary">
              {__("Select approvers to request approval before publishing, or publish directly without approvers.")}
            </p>
            <PeopleMultiSelectField
              name="approverIds"
              label={__("Approvers")}
              control={control}
              organizationId={organizationId}
              placeholder={__("Add approvers...")}
            />
          </div>
        </DialogContent>
        <DialogFooter>
          <Button
            type="submit"
            icon={hasApprovers ? IconSend : IconUpload}
            disabled={isPublishing}
          >
            {hasApprovers ? __("Request approval") : __("Publish")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
