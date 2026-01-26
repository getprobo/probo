import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
} from "@probo/ui";
import { type ReactNode } from "react";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { PublishDocumentsDialogMutation } from "#/__generated__/core/PublishDocumentsDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

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
          ...DocumentListItemFragment
        }
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
  const dialogRef = useDialogRef();

  const schema = z.object({
    changelog: z.string().min(1, __("Changelog is required")),
  });

  const [publishMutation]
    = useMutationWithToasts<PublishDocumentsDialogMutation>(
      documentsPublishMutation,
      {
        successMessage: (response) => {
          const actualPublishedCount
            = response.bulkPublishDocumentVersions.documentEdges.length;
          return sprintf(__("%s documents published"), actualPublishedCount);
        },
        errorMessage: sprintf(
          __("Failed to publish %s documents"),
          documentIds.length,
        ),
      },
    );

  const {
    handleSubmit,
    register,
    formState: { isSubmitting, errors },
  } = useFormWithSchema(schema, {
    defaultValues: {
      changelog: "",
    },
  });

  const onSubmit = async (data: z.infer<typeof schema>) => {
    await publishMutation({
      variables: {
        input: {
          documentIds,
          changelog: data.changelog,
        },
      },
      onSuccess: () => {
        dialogRef.current?.close();
        onSave();
      },
    });
  };

  return (
    <Dialog
      className="max-w-xl"
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Documents"), __("Publish documents")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
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
