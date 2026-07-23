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
import { useMemo, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { PublishObligationListDialogMutation } from "#/__generated__/core/PublishObligationListDialogMutation.graphql";
import { PeopleMultiSelectField } from "#/components/form/PeopleMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const publishMutation = graphql`
  mutation PublishObligationListDialogMutation(
    $input: PublishObligationListInput!
  ) {
    publishObligationList(input: $input) {
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
  organizationId: string;
  defaultApproverIds?: string[];
  onPublished?: (documentId: string) => void;
};

export function PublishObligationListDialog({
  children,
  organizationId,
  defaultApproverIds,
  onPublished,
}: Props) {
  const { t } = useTranslation();
  const { toast } = useToast();
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
      approverIds: defaultApproverIds ?? [],
    },
  });

  const [publish, isPublishing]
    = useMutation<PublishObligationListDialogMutation>(publishMutation);

  const minorRef = useRef(false);

  const approverIds = watch("approverIds");
  const hasApprovers = approverIds.length > 0;

  const onSubmit = (data: z.infer<typeof schema>) => {
    publish({
      variables: {
        input: {
          minor: minorRef.current,
          organizationId,
          approverIds: !minorRef.current && data.approverIds.length > 0 ? data.approverIds : undefined,
        },
      },
      onCompleted(response) {
        const documentId = response.publishObligationList?.documentEdge?.node?.id;
        if (documentId) {
          toast({
            title: t("publishObligationListDialog.messages.success"),
            description: hasApprovers
              ? t("publishObligationListDialog.messages.approvalRequested")
              : t("publishObligationListDialog.messages.published"),
            variant: "success",
          });
          dialogRef.current?.close();
          reset();
          onPublished?.(documentId);
        }
      },
      onError(error) {
        toast({
          title: t("publishObligationListDialog.messages.error"),
          description: formatError(
            t("publishObligationListDialog.errors.publish"),
            error,
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
      title={t("publishObligationListDialog.title")}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded>
          <div className="space-y-4">
            <p className="text-sm text-txt-secondary">
              {t("publishObligationListDialog.description")}
            </p>
            <PeopleMultiSelectField
              name="approverIds"
              label={t("publishObligationListDialog.fields.approvers")}
              control={control}
              organizationId={organizationId}
              placeholder={t("publishObligationListDialog.fields.approversPlaceholder")}
            />
          </div>
        </DialogContent>
        <DialogFooter>
          <Button
            type="submit"
            variant="secondary"
            icon={IconUpload}
            onClick={() => { minorRef.current = true; }}
            disabled={isPublishing}
          >
            {t("publishObligationListDialog.actions.publishMinor")}
          </Button>
          <Button
            type="submit"
            icon={hasApprovers ? IconSend : IconUpload}
            onClick={() => { minorRef.current = false; }}
            disabled={isPublishing}
          >
            {hasApprovers ? t("publishObligationListDialog.actions.requestApproval") : t("publishObligationListDialog.actions.publish")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
