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
  IconUpload,
  IconWarning,
  Textarea,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode, useRef } from "react";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const minorRef = useRef(false);

  const schema = z.object({
    changelog: z.string().min(1, t("publishDocumentsDialog.validation.changelogRequired")),
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
            title: t("publishDocumentsDialog.errors.title"),
            description: formatError(t("publishDocumentsDialog.errors.publish"), [...errors]),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("publishDocumentsDialog.messages.successTitle"),
          description: t("publishDocumentsDialog.messages.published", {
            count: documentIds.length,
          }),
          variant: "success",
        });
        dialogRef.current?.close();
        onSave();
      },
      onError(error) {
        toast({
          title: t("publishDocumentsDialog.errors.title"),
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
      title={t("publishDocumentsDialog.title")}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded>
          <div className="space-y-4">
            <div className="flex items-start gap-2 rounded-lg bg-bg-warning/10 border border-border-warning p-3">
              <IconWarning size={16} className="text-txt-warning shrink-0 mt-0.5" />
              <div className="text-sm text-txt-warning space-y-1">
                <p>{t("publishDocumentsDialog.warning.approval")}</p>
                <p>{t("publishDocumentsDialog.warning.skipped")}</p>
              </div>
            </div>
            <div>
              <label htmlFor="changelog" className="text-sm font-medium text-txt-primary mb-1 block">
                {t("publishDocumentsDialog.fields.changelog")}
              </label>
              <Textarea
                id="changelog"
                aria-label={t("publishDocumentsDialog.fields.changelog")}
                required
                autogrow
                placeholder={t("publishDocumentsDialog.fields.changelogPlaceholder")}
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
            {t("publishDocumentsDialog.actions.publishMinor")}
          </Button>
          <Button
            type="submit"
            disabled={isPublishing}
            onClick={() => { minorRef.current = false; }}
          >
            {t("publishDocumentsDialog.actions.publish", { count: documentIds.length })}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
