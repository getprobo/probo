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

import { Button, Dialog, DialogContent, DialogFooter, type DialogRef, Field, IconCircleInfo, Spinner, Textarea } from "@probo/ui";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { ComplianceUpdateFormDialogCreateMutation } from "#/__generated__/core/ComplianceUpdateFormDialogCreateMutation.graphql";
import type { ComplianceUpdateFormDialogUpdateMutation } from "#/__generated__/core/ComplianceUpdateFormDialogUpdateMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutation } from "#/lib/relay/useMutation";

import type { UpdateNode } from "./CompliancePageUpdatesList";

const createMutation = graphql`
  mutation ComplianceUpdateFormDialogCreateMutation(
    $input: CreateMailingListUpdateInput!
    $connections: [ID!]!
  ) {
    createMailingListUpdate(input: $input) {
      mailingListUpdate
        @prependNode(connections: $connections, edgeTypeName: "MailingListUpdateEdge") {
        id
        title
        body
        status
        createdAt
        updatedAt
      }
    }
  }
`;

const updateMutation = graphql`
  mutation ComplianceUpdateFormDialogUpdateMutation($input: UpdateMailingListUpdateInput!) {
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

type CreateProps = {
  ref: DialogRef;
  mailingListId: string;
  connectionId: string;
  onCreated?: () => void;
  update?: never;
};

type EditProps = {
  ref: DialogRef;
  update: UpdateNode | null;
  mailingListId?: never;
  connectionId?: never;
  onCreated?: never;
};

type Props = CreateProps | EditProps;

export function ComplianceUpdateFormDialog(props: Props) {
  const { ref, update, mailingListId, connectionId, onCreated } = props;
  const { t } = useTranslation("organizations/compliance-page");

  const isEditMode = update !== undefined;
  const isSent = isEditMode && update?.status !== "DRAFT";

  const schemaWithMessages = z.object({
    title: z.string().trim().min(1, t("updateFormDialog.validation.titleRequired")),
    body: z.string().trim().min(1, t("updateFormDialog.validation.bodyRequired")),
  });

  const form = useFormWithSchema(schemaWithMessages, {
    defaultValues: { title: "", body: "" },
  });

  useEffect(() => {
    if (update) {
      form.reset({ title: update.title, body: update.body });
    }
  }, [update, form]);

  const [createUpdate, isCreating] = useMutation<ComplianceUpdateFormDialogCreateMutation>(
    createMutation,
    {
      successMessage: t("updateFormDialog.messages.created"),
      errorToast: t("updateFormDialog.errors.create"),
    },
  );

  const [saveUpdate, isSaving] = useMutation<ComplianceUpdateFormDialogUpdateMutation>(
    updateMutation,
    {
      successMessage: t("updateFormDialog.messages.updated"),
      errorToast: t("updateFormDialog.errors.update"),
    },
  );

  const handleSubmit = async (data: z.infer<typeof schemaWithMessages>) => {
    if (isEditMode) {
      if (!update) return;
      await saveUpdate({
        variables: {
          input: {
            id: update.id,
            title: data.title,
            body: data.body,
          },
        },
        onCompleted: (_, errors) => {
          if (!errors?.length) {
            ref.current?.close();
          }
        },
      });
    } else {
      await createUpdate({
        variables: {
          input: {
            mailingListId: mailingListId,
            title: data.title,
            body: data.body,
          },
          connections: [connectionId],
        },
        onCompleted: (_, errors) => {
          if (!errors?.length) {
            form.reset();
            ref.current?.close();
            onCreated?.();
          }
        },
      });
    }
  };

  return (
    <Dialog ref={ref} title={isSent ? t("updateFormDialog.title.view") : isEditMode ? t("updateFormDialog.title.edit") : t("updateFormDialog.title.create")}>
      <form onSubmit={e => void form.handleSubmit(handleSubmit)(e)}>
        <DialogContent className="px-6 pt-4 pb-2 space-y-4">
          <div className="flex gap-2.5 rounded-lg bg-surface-secondary p-3 text-sm text-txt-secondary">
            <IconCircleInfo size={16} className="mt-0.5 shrink-0 text-txt-tertiary" />
            {t("updateFormDialog.confidentialityNotice")}
          </div>
          <Field
            label={t("updateFormDialog.fields.title")}
            required
            disabled={isSent}
            error={form.formState.errors.title?.message}
            {...form.register("title")}
          />
          <Field
            label={t("updateFormDialog.fields.body")}
            required
            error={form.formState.errors.body?.message}
          >
            <Textarea rows={12} disabled={isSent} {...form.register("body")} />
          </Field>
          {isSent && (
            <p className="text-sm text-txt-tertiary">
              {t("updateFormDialog.sentNotice")}
            </p>
          )}
        </DialogContent>
        <DialogFooter>
          {!isSent && (
            <Button type="submit" disabled={isCreating || isSaving}>
              {(isCreating || isSaving) && <Spinner />}
              {isEditMode ? t("updateFormDialog.actions.save") : t("updateFormDialog.actions.create")}
            </Button>
          )}
        </DialogFooter>
      </form>
    </Dialog>
  );
}
