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

import { Button, Checkbox, Dialog, DialogContent, DialogFooter, type DialogRef, Field, Spinner } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { type DataID, graphql } from "relay-runtime";
import { z } from "zod";

import type { NewCompliancePageSubscriberDialogMutation } from "#/__generated__/core/NewCompliancePageSubscriberDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutation } from "#/lib/relay/useMutation";

const createSubscriberMutation = graphql`
  mutation NewCompliancePageSubscriberDialogMutation(
    $input: CreateMailingListSubscriberInput!
    $connections: [ID!]!
  ) {
    createMailingListSubscriber(input: $input) {
      mailingListSubscriberEdge @prependEdge(connections: $connections) {
        cursor
        node {
          id
          fullName
          email
          status
          createdAt
        }
      }
    }
  }
`;

export function NewCompliancePageSubscriberDialog(props: {
  mailingListId: string;
  connectionId: DataID;
  ref: DialogRef;
}) {
  const { mailingListId, connectionId, ref } = props;
  const { t } = useTranslation("organizations/compliance-page");

  const schema = z.object({
    fullName: z.string().trim().min(1, t("newSubscriberDialog.validation.fullNameRequired")),
    email: z
      .string()
      .min(1, t("newSubscriberDialog.validation.emailRequired"))
      .trim()
      .email(t("newSubscriberDialog.validation.emailInvalid")),
    confirmed: z.boolean(),
  });

  const form = useFormWithSchema(schema, {
    defaultValues: { fullName: "", email: "", confirmed: false },
  });

  const [createSubscriber, isCreating] = useMutation<NewCompliancePageSubscriberDialogMutation>(
    createSubscriberMutation,
    {
      successMessage: t("newSubscriberDialog.messages.created"),
      errorToast: t("newSubscriberDialog.errors.create"),
    },
  );

  const handleSubmit = async (data: z.infer<typeof schema>) => {
    await createSubscriber({
      variables: {
        input: {
          mailingListId,
          fullName: data.fullName.trim(),
          email: data.email.trim(),
          confirmed: data.confirmed || undefined,
        },
        connections: connectionId ? [connectionId] : [],
      },
      onCompleted: (_, errors) => {
        if (errors?.length) return;
        setTimeout(() => {
          form.reset();
          ref.current?.close();
        }, 50);
        setTimeout(() => {
          form.reset();
        }, 300);
      },
    });
  };

  return (
    <Dialog ref={ref} title={t("newSubscriberDialog.title")}>
      <form onSubmit={e => void form.handleSubmit(handleSubmit)(e)}>
        <DialogContent padded className="space-y-6">
          <p className="text-txt-secondary text-sm">
            {t("newSubscriberDialog.description")}
          </p>
          <Field
            label={t("newSubscriberDialog.fields.fullName")}
            required
            error={form.formState.errors.fullName?.message}
            type="text"
            {...form.register("fullName")}
            placeholder={t("newSubscriberDialog.fields.fullNamePlaceholder")}
          />
          <Field
            label={t("newSubscriberDialog.fields.email")}
            required
            error={form.formState.errors.email?.message}
            type="email"
            {...form.register("email")}
            placeholder={t("newSubscriberDialog.fields.emailPlaceholder")}
          />
          <div className="space-y-2">
            <label className="flex items-center gap-2 cursor-pointer">
              <Checkbox
                checked={form.watch("confirmed")}
                onChange={checked => form.setValue("confirmed", checked)}
              />
              <span className="text-sm font-medium">
                {t("newSubscriberDialog.fields.skipConfirmation")}
              </span>
            </label>
            {form.watch("confirmed") && (
              <p className="text-txt-secondary text-xs pl-6">
                {t("newSubscriberDialog.consentNotice")}
              </p>
            )}
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isCreating}>
            {isCreating && <Spinner />}
            {t("newSubscriberDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
