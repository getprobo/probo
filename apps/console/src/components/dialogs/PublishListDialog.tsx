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

import { PeopleMultiSelectField } from "#/components/form/PeopleMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { z } from "#/lib/zod";

export type PublishListDialogLabels = {
  title: string;
  description: string;
  approvers: string;
  approversPlaceholder: string;
  publishMinor: string;
  publish: string;
  requestApproval: string;
  successTitle: string;
  published: string;
  approvalRequested: string;
  errorTitle: string;
  publishError: string;
};

export type PublishListDialogInput = {
  organizationId: string;
  minor: boolean;
  approverIds?: string[];
};

type Props = {
  children: ReactNode;
  organizationId: string;
  defaultApproverIds?: string[];
  labels: PublishListDialogLabels;
  isPublishing: boolean;
  onPublish: (input: PublishListDialogInput) => Promise<string | null | undefined>;
  onPublished?: (documentId: string) => void;
};

export function PublishListDialog({
  children,
  organizationId,
  defaultApproverIds,
  labels,
  isPublishing,
  onPublish,
  onPublished,
}: Props) {
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
        title: labels.successTitle,
        description: requestedApproval
          ? labels.approvalRequested
          : labels.published,
        variant: "success",
      });
      dialogRef.current?.close();
      reset();
      onPublished?.(documentId);
    } catch (error) {
      toast({
        title: labels.errorTitle,
        description: formatError(
          labels.publishError,
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
      title={labels.title}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded>
          <div className="space-y-4">
            <p className="text-sm text-txt-secondary">
              {labels.description}
            </p>
            <PeopleMultiSelectField
              name="approverIds"
              label={labels.approvers}
              control={control}
              organizationId={organizationId}
              placeholder={labels.approversPlaceholder}
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
            {labels.publishMinor}
          </Button>
          <Button
            type="submit"
            icon={hasApprovers ? IconSend : IconUpload}
            onClick={() => { minorRef.current = false; }}
            disabled={isPublishing}
          >
            {hasApprovers
              ? labels.requestApproval
              : labels.publish}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
