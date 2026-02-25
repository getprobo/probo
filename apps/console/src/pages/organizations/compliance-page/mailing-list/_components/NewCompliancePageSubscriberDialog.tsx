import { useTranslate } from "@probo/i18n";
import { Button, Dialog, DialogContent, DialogFooter, type DialogRef, Field, Spinner } from "@probo/ui";
import { type DataID, graphql } from "relay-runtime";
import { z } from "zod";

import type { NewCompliancePageSubscriberDialogMutation } from "#/__generated__/core/NewCompliancePageSubscriberDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

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
  const { __ } = useTranslate();

  const schema = z.object({
    fullName: z.string().min(1, __("Full name is required")).trim(),
    email: z
      .string()
      .min(1, __("Email is required"))
      .trim()
      .email(__("Please enter a valid email address")),
  });

  const form = useFormWithSchema(schema, {
    defaultValues: { fullName: "", email: "" },
  });

  const [createSubscriber, isCreating] = useMutationWithToasts<NewCompliancePageSubscriberDialogMutation>(
    createSubscriberMutation,
    {
      successMessage: __("Subscriber added successfully"),
      errorMessage: __("Failed to add subscriber"),
    },
  );

  const handleSubmit = async (data: z.infer<typeof schema>) => {
    await createSubscriber({
      variables: {
        input: {
          mailingListId,
          fullName: data.fullName.trim(),
          email: data.email.trim(),
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
    <Dialog ref={ref} title={__("Add Subscriber")}>
      <form onSubmit={e => void form.handleSubmit(handleSubmit)(e)}>
        <DialogContent padded className="space-y-6">
          <p className="text-txt-secondary text-sm">
            {__("Add a person to receive security and compliance updates")}
          </p>
          <Field
            label={__("Full Name")}
            required
            error={form.formState.errors.fullName?.message}
            {...form.register("fullName")}
            placeholder={__("John Doe")}
          />
          <Field
            label={__("Email Address")}
            required
            error={form.formState.errors.email?.message}
            type="email"
            {...form.register("email")}
            placeholder={__("john@example.com")}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isCreating}>
            {isCreating && <Spinner />}
            {__("Add Subscriber")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
