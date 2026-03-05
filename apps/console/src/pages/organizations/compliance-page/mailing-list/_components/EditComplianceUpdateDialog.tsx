import { useTranslate } from "@probo/i18n";
import { Button, Dialog, DialogContent, DialogFooter, type DialogRef, Field, IconCircleInfo, Spinner, Textarea } from "@probo/ui";
import { useEffect } from "react";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { EditComplianceUpdateDialogMutation } from "#/__generated__/core/EditComplianceUpdateDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import type { UpdateNode } from "./CompliancePageUpdatesList";

const updateMutation = graphql`
  mutation EditComplianceUpdateDialogMutation($input: UpdateMailingListUpdateInput!) {
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

export function EditComplianceUpdateDialog(props: {
  update: UpdateNode | null;
  ref: DialogRef;
}) {
  const { update, ref } = props;
  const { __ } = useTranslate();

  const isSent = update?.status !== "DRAFT";

  const schema = z.object({
    title: z.string().trim().min(1, __("Title is required")),
    body: z.string().trim().min(1, __("Body is required")),
  });

  const form = useFormWithSchema(schema, {
    defaultValues: { title: "", body: "" },
  });

  useEffect(() => {
    if (update) {
      form.reset({ title: update.title, body: update.body });
    }
  }, [update, form]);

  const [saveUpdate, isSaving] = useMutationWithToasts<EditComplianceUpdateDialogMutation>(
    updateMutation,
    {
      successMessage: __("Update saved successfully"),
      errorMessage: __("Failed to save update"),
    },
  );

  const handleSubmit = async (data: { title: string; body: string }) => {
    if (!update) return;
    await saveUpdate({
      variables: {
        input: {
          id: update.id,
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
          <div className="flex gap-2.5 rounded-lg bg-surface-secondary p-3 text-sm text-txt-secondary">
            <IconCircleInfo size={16} className="mt-0.5 shrink-0 text-txt-tertiary" />
            <p>
              {__("Do not include confidential or sensitive information in this update. Any content that requires protection should be placed behind your NDA-gated documents instead.")}
            </p>
          </div>
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
            <Button type="submit" disabled={isSaving}>
              {isSaving && <Spinner />}
              {__("Save")}
            </Button>
          )}
        </DialogFooter>
      </form>
    </Dialog>
  );
}
