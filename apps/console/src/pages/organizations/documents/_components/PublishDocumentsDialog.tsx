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

import { formatError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconUpload,
  IconWarning,
  Textarea,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode, useRef } from "react";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { PublishDocumentsDialog_majorMutation } from "#/__generated__/core/PublishDocumentsDialog_majorMutation.graphql";
import type { PublishDocumentsDialog_minorMutation } from "#/__generated__/core/PublishDocumentsDialog_minorMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

type Props = {
  documentIds: string[];
  children: ReactNode;
  onSave: () => void;
};

const publishMajorMutation = graphql`
  mutation PublishDocumentsDialog_majorMutation(
    $input: BulkPublishDocumentVersionsInput!
  ) {
    bulkPublishMajorDocumentVersions(input: $input) {
      documentVersions {
        id
      }
      documents {
        id
        ...DocumentListItemFragment
      }
    }
  }
`;

const publishMinorMutation = graphql`
  mutation PublishDocumentsDialog_minorMutation(
    $input: BulkPublishDocumentVersionsInput!
  ) {
    bulkPublishMinorDocumentVersions(input: $input) {
      documentVersions {
        id
      }
      documents {
        id
        ...DocumentListItemFragment
      }
    }
  }
`;

export function PublishDocumentsDialog({
  documentIds,
  children,
  onSave,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const actionRef = useRef<"major" | "minor">("major");

  const schema = z.object({
    changelog: z.string().min(1, __("Changelog is required")),
  });

  const [publishMajor, isPublishingMajor]
    = useMutation<PublishDocumentsDialog_majorMutation>(publishMajorMutation);
  const [publishMinor, isPublishingMinor]
    = useMutation<PublishDocumentsDialog_minorMutation>(publishMinorMutation);

  const isBusy = isPublishingMajor || isPublishingMinor;

  const {
    handleSubmit,
    register,
    formState: { errors },
  } = useFormWithSchema(schema, {
    defaultValues: {
      changelog: "",
    },
  });

  const onCompleted = (_: unknown, errors: ReadonlyArray<{ message: string }> | null) => {
    if (errors?.length) {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to publish documents"), [...errors]),
        variant: "error",
      });
    } else {
      toast({
        title: __("Success"),
        description: sprintf(__("%s documents published"), documentIds.length),
        variant: "success",
      });
      dialogRef.current?.close();
      onSave();
    }
  };

  const onError = (error: Error) => {
    toast({
      title: __("Error"),
      description: error.message,
      variant: "error",
    });
  };

  const onSubmit = (data: z.infer<typeof schema>) => {
    const variables = {
      input: {
        documentIds,
        changelog: data.changelog,
      },
    };

    if (actionRef.current === "minor") {
      publishMinor({ variables, onCompleted, onError });
    } else {
      publishMajor({ variables, onCompleted, onError });
    }
  };

  return (
    <Dialog
      className="max-w-xl"
      ref={dialogRef}
      trigger={children}
      title={__("Publish documents")}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded>
          <div className="space-y-4">
            <div className="flex items-start gap-2 rounded-lg bg-bg-warning/10 border border-border-warning p-3">
              <IconWarning size={16} className="text-txt-warning shrink-0 mt-0.5" />
              <div className="text-sm text-txt-warning space-y-1">
                <p>{__("Publishing as major will request approval for documents that have default approvers configured. Approvers will receive an email notification.")}</p>
                <p>{__("Documents already published and pending approval will be skipped.")}</p>
              </div>
            </div>
            <div>
              <label htmlFor="changelog" className="text-sm font-medium text-txt-primary mb-1 block">
                {__("Changelog")}
              </label>
              <Textarea
                id="changelog"
                aria-label={__("Changelog")}
                required
                autogrow
                placeholder={__("Describe what changed in this version...")}
                {...register("changelog")}
              />
              {errors.changelog?.message && (
                <p className="text-xs text-txt-danger mt-1">{errors.changelog.message}</p>
              )}
            </div>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button
            type="submit"
            variant="secondary"
            icon={IconUpload}
            disabled={isBusy}
            onClick={() => { actionRef.current = "minor"; }}
          >
            {__("Publish as minor")}
          </Button>
          <Button
            type="submit"
            disabled={isBusy}
            onClick={() => { actionRef.current = "major"; }}
          >
            {sprintf(__("Publish %s documents"), documentIds.length)}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
