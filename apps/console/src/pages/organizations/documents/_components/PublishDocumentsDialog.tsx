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

import type { PublishDocumentsDialog_bulkPublishMutation } from "#/__generated__/core/PublishDocumentsDialog_bulkPublishMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

type Props = {
  documentIds: string[];
  children: ReactNode;
  onSave: () => void;
};

const bulkPublishMutation = graphql`
  mutation PublishDocumentsDialog_bulkPublishMutation(
    $input: BulkPublishDocumentsInput!
  ) {
    bulkPublishDocuments(input: $input) {
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
  const minorRef = useRef(false);

  const schema = z.object({
    changelog: z.string().min(1, __("Changelog is required")),
  });

  const [publish, isPublishing]
    = useMutation<PublishDocumentsDialog_bulkPublishMutation>(bulkPublishMutation);

  const {
    handleSubmit,
    register,
    formState: { errors },
  } = useFormWithSchema(schema, {
    defaultValues: {
      changelog: "",
    },
  });

  const onSubmit = (data: z.infer<typeof schema>) => {
    const minor = minorRef.current;
    minorRef.current = false;
    publish({
      variables: {
        input: {
          documentIds,
          minor,
          changelog: data.changelog,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Failed to publish documents"), [...errors]),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: sprintf(__("%s documents published"), documentIds.length),
          variant: "success",
        });
        dialogRef.current?.close();
        onSave();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: error.message,
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
            disabled={isPublishing}
            onClick={() => { minorRef.current = true; }}
          >
            {__("Publish as minor")}
          </Button>
          <Button
            type="submit"
            disabled={isPublishing}
            onClick={() => { minorRef.current = false; }}
          >
            {sprintf(__("Publish %s documents"), documentIds.length)}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
