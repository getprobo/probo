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

import { PeopleMultiSelectField } from "#/components/form/PeopleMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { z } from "#/lib/zod";

export type PublishListDialogInput = {
  organizationId: string;
  minor: boolean;
  approverIds?: string[];
};

type Props = {
  children: ReactNode;
  organizationId: string;
  defaultApproverIds?: string[];
  title: string;
  publishedMessage: string;
  publishError: string;
  isPublishing: boolean;
  onPublish: (input: PublishListDialogInput) => Promise<string | null | undefined>;
  onPublished?: (documentId: string) => void;
};

export function PublishListDialog({
  children,
  organizationId,
  defaultApproverIds,
  title,
  publishedMessage,
  publishError,
  isPublishing,
  onPublish,
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

  const minorRef = useRef(false);

  const approverIds = watch("approverIds");
  const hasApprovers = approverIds.length > 0;

  const onSubmit = async (data: z.infer<typeof schema>) => {
    const requestedApproval = !minorRef.current && data.approverIds.length > 0;

    try {
      const documentId = await onPublish({
        organizationId,
        minor: minorRef.current,
        approverIds: requestedApproval ? data.approverIds : undefined,
      });

      if (!documentId) {
        return;
      }

      toast({
        title: t("publishListDialog.messages.success"),
        description: requestedApproval
          ? t("publishListDialog.messages.approvalRequested")
          : publishedMessage,
        variant: "success",
      });
      dialogRef.current?.close();
      reset();
      onPublished?.(documentId);
    } catch (error) {
      toast({
        title: t("publishListDialog.messages.error"),
        description: formatError(
          publishError,
          error as Parameters<typeof formatError>[1],
        ),
        variant: "error",
      });
    }
  };

  return (
    <Dialog
      className="max-w-xl"
      ref={dialogRef}
      trigger={children}
      title={title}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded>
          <div className="space-y-4">
            <p className="text-sm text-txt-secondary">
              {t("publishListDialog.description")}
            </p>
            <PeopleMultiSelectField
              name="approverIds"
              label={t("publishListDialog.fields.approvers")}
              control={control}
              organizationId={organizationId}
              placeholder={t("publishListDialog.fields.approversPlaceholder")}
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
            {t("publishListDialog.actions.publishMinor")}
          </Button>
          <Button
            type="submit"
            icon={hasApprovers ? IconSend : IconUpload}
            onClick={() => { minorRef.current = false; }}
            disabled={isPublishing}
          >
            {hasApprovers
              ? t("publishListDialog.actions.requestApproval")
              : t("publishListDialog.actions.publish")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
