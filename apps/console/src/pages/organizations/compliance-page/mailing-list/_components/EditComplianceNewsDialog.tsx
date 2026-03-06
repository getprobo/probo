import { useTranslate } from "@probo/i18n";
import { Button, Dialog, DialogContent, DialogFooter, type DialogRef, Field, Spinner, Textarea } from "@probo/ui";
import { useEffect } from "react";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { EditComplianceNewsDialogMutation } from "#/__generated__/core/EditComplianceNewsDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const updateMutation = graphql`
  mutation EditComplianceNewsDialogMutation($input: UpdateMailingListUpdateInput!) {
    updateMailingListUpdate(input: $input) {
      mailingListUpdate {
        id
        title
        body
        status
        updatedAt
      }
    }
  }
`;

type NewsItem = {
  id: string;
  title: string;
  body: string;
  status: "DRAFT" | "ENQUEUED" | "PROCESSING" | "SENT";
};

export function EditComplianceNewsDialog(props: {
  news: NewsItem | null;
  ref: DialogRef;
}) {
  const { news, ref } = props;
  const { __ } = useTranslate();

  const isSent = news?.status !== "DRAFT";

  const schema = z.object({
    title: z.string().trim().min(1, __("Title is required")),
    body: z.string().trim().min(1, __("Body is required")),
  });

  const form = useFormWithSchema(schema, {
    defaultValues: { title: "", body: "" },
  });

  useEffect(() => {
    if (news) {
      form.reset({ title: news.title, body: news.body });
    }
  }, [news, form]);

  const [updateNews, isUpdating] = useMutationWithToasts<EditComplianceNewsDialogMutation>(
    updateMutation,
    {
      successMessage: __("News updated successfully"),
      errorMessage: __("Failed to update news"),
    },
  );

  const handleSubmit = async (data: { title: string; body: string }) => {
    if (!news) return;
    await updateNews({
      variables: {
        input: {
          id: news.id,
          title: data.title.trim(),
          body: data.body.trim(),
        },
      },
      onCompleted: (_, errors) => {
        if (!errors?.length) {
          ref.current?.close();
        }
      },
    });
  };

  return (
    <Dialog ref={ref} title={__("Edit Update")}>
      <form onSubmit={e => void form.handleSubmit(handleSubmit)(e)}>
        <DialogContent padded className="space-y-6">
          <Field
            label={__("Title")}
            required
            disabled={isSent}
            error={form.formState.errors.title?.message}
            {...form.register("title")}
          />
          <Field
            label={__("Body")}
            required
            error={form.formState.errors.body?.message}
          >
            <Textarea rows={6} disabled={isSent} {...form.register("body")} />
          </Field>
          {isSent && (
            <p className="text-sm text-txt-tertiary">
              {__("This update has been sent and can no longer be edited.")}
            </p>
          )}
        </DialogContent>
        <DialogFooter>
          {!isSent && (
            <Button type="submit" disabled={isUpdating}>
              {isUpdating && <Spinner />}
              {__("Save")}
            </Button>
          )}
        </DialogFooter>
      </form>
    </Dialog>
  );
}
