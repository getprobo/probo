import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
  Breadcrumb,
} from "@probo/ui";
import { type ReactNode } from "react";
import { useTranslate } from "@probo/i18n";
import z from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema.ts";
import { graphql } from "relay-runtime";
import { sprintf } from "@probo/helpers";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts.ts";

type Props = {
  documentIds: string[];
  children: ReactNode;
  onSave: () => void;
};

const documentsPublishMutation = graphql`
  mutation PublishDocumentsDialogMutation(
    $input: BulkPublishDocumentVersionsInput!
  ) {
    bulkPublishDocumentVersions(input: $input) {
      documentEdges {
        node {
          id
          ...DocumentsPageRowFragment
        }
      }
    }
  }
`;

const schema = z.object({
  changelog: z.string().min(1, "Changelog is required"),
});

export function PublishDocumentsDialog({
  documentIds,
  children,
  onSave,
}: Props) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const {
    handleSubmit,
    register,
    formState: { isSubmitting, errors },
  } = useFormWithSchema(schema, {
    defaultValues: {
      changelog: "",
    },
  });
  const onSubmit = handleSubmit(async (data) => {
    await publishMutation({
      variables: {
        input: {
          documentIds,
          changelog: data.changelog,
        },
      },
    });
    dialogRef.current?.close();
    onSave();
  });

  const [publishMutation] = useMutationWithToasts(documentsPublishMutation, {
    successMessage: sprintf(__("%s documents published"), documentIds.length),
    errorMessage: sprintf(
      __("Failed to publish %s documents"),
      documentIds.length
    ),
  });

  return (
    <Dialog
      className="max-w-xl"
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Documents"), __("Publish documents")]} />}
    >
      <form onSubmit={onSubmit}>
        <DialogContent padded>
          <Field
            id="changelog"
            aria-label={__("Changelog")}
            required
            variant="title"
            placeholder={__("Changelog")}
            {...register("changelog")}
            error={errors.changelog?.message}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isSubmitting}>
            {sprintf(__("Publish %s documents"), documentIds.length)}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
