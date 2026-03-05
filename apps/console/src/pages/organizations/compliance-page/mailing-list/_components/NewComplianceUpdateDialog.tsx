import { useTranslate } from "@probo/i18n";
import { Button, Dialog, DialogContent, DialogFooter, type DialogRef, Field, Spinner, Textarea } from "@probo/ui";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { NewComplianceUpdateDialogMutation } from "#/__generated__/core/NewComplianceUpdateDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const createMutation = graphql`
  mutation NewComplianceUpdateDialogMutation(
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

export function NewComplianceUpdateDialog(props: {
  mailingListId: string;
  connectionId: string;
  ref: DialogRef;
  onCreated?: () => void;
}) {
  const { mailingListId, connectionId, ref, onCreated } = props;
  const { __ } = useTranslate();

  const schema = z.object({
    title: z.string().trim().min(1, __("Title is required")),
    body: z.string().trim().min(1, __("Body is required")),
  });

  const form = useFormWithSchema(schema, {
    defaultValues: { title: "", body: "" },
  });

  const [createUpdate, isCreating] = useMutationWithToasts<NewComplianceUpdateDialogMutation>(
    createMutation,
    {
      successMessage: __("Update created successfully"),
      errorMessage: __("Failed to create update"),
    },
  );

  const handleSubmit = async (data: { title: string; body: string }) => {
    await createUpdate({
      variables: {
        input: {
          mailingListId,
          title: data.title.trim(),
          body: data.body.trim(),
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
  };

  return (
    <Dialog ref={ref} title={__("Add Update")}>
      <form onSubmit={e => void form.handleSubmit(handleSubmit)(e)}>
        <DialogContent padded className="space-y-6">
          <Field
            label={__("Title")}
            required
            error={form.formState.errors.title?.message}
            {...form.register("title")}
          />
          <Field
            label={__("Body")}
            required
            error={form.formState.errors.body?.message}
          >
            <Textarea rows={6} {...form.register("body")} />
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isCreating}>
            {isCreating && <Spinner />}
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
