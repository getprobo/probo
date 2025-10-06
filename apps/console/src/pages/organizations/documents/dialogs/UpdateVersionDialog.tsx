import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Spinner,
  Textarea,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type RefObject } from "react";
import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import type { UpdateVersionDialogCreateMutation } from "./__generated__/UpdateVersionDialogCreateMutation.graphql";
import type { UpdateVersionDialogUpdateMutation } from "./__generated__/UpdateVersionDialogUpdateMutation.graphql";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type { DocumentDetailPageDocumentFragment$data } from "../__generated__/DocumentDetailPageDocumentFragment.graphql";

const createDraftDocument = graphql`
  mutation UpdateVersionDialogCreateMutation(
    $input: CreateDraftDocumentVersionInput!
    $connections: [ID!]!
  ) {
    createDraftDocumentVersion(input: $input) {
      documentVersionEdge @prependEdge(connections: $connections) {
        node {
          id
          content
          status
          publishedAt
          version
          updatedAt
          signatures(first: 100) {
            edges {
              node {
                id
                state
              }
            }
          }
        }
      }
    }
  }
`;

const UpdateDocumentMutation = graphql`
  mutation UpdateVersionDialogUpdateMutation(
    $input: UpdateDocumentVersionInput!
  ) {
    updateDocumentVersion(input: $input) {
      documentVersion {
        id
        content
      }
    }
  }
`;

type Props = {
  document: DocumentDetailPageDocumentFragment$data;
  connectionId: string;
  ref: RefObject<{ open: () => void } | null>;
};

const versionSchema = z.object({
  content: z.string(),
});

export default function UpdateVersionDialog({
  document,
  connectionId,
  ref,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const version = document.versions.edges[0].node;
  const isDraft = version?.status === "DRAFT";
  const [createDraftDocumentVersion, isCreatingDraft] =
    useMutation<UpdateVersionDialogCreateMutation>(createDraftDocument);
  const [updateDocumentVersion, isUpdating] =
    useMutationWithToasts<UpdateVersionDialogUpdateMutation>(
      UpdateDocumentMutation,
      {
        successMessage: __("Document updated successfully."),
        errorMessage: __("Failed to update document"),
      }
    );
  const { handleSubmit, register } = useFormWithSchema(versionSchema, {
    defaultValues: {
      content: version.content,
    },
  });

  ref.current = {
    open: () => {
      dialogRef.current?.open();
    },
  };

  if (!version) {
    return;
  }

  const onSubmit = handleSubmit((data) => {
    if (isDraft) {
      updateDocumentVersion({
        variables: {
          input: {
            documentVersionId: version.id,
            content: data.content,
          },
        },
        onSuccess: () => {
          dialogRef.current?.close();
        },
      });
    } else {
      createDraftDocumentVersion({
        variables: {
          input: {
            documentID: document.id,
          },
          connections: [connectionId],
        },
        onCompleted: (createResponse, errors) => {
          if (errors) {
            toast({
              variant: "error",
              title: __("Error creating draft"),
              description:
                errors[0]?.message || __("An unknown error occurred"),
            });
            return;
          }

          const newVersionId = createResponse?.createDraftDocumentVersion?.documentVersionEdge?.node?.id;
          if (newVersionId && data.content !== version.content) {
            updateDocumentVersion({
              variables: {
                input: {
                  documentVersionId: newVersionId,
                  content: data.content,
                },
              },
              onSuccess: () => {
                dialogRef.current?.close();
              },
            });
          } else {
            dialogRef.current?.close();
          }
        },
      });
    }
  });

  const isLoading = isCreatingDraft || isUpdating;

  return (
    <Dialog
      ref={dialogRef}
      title={<Breadcrumb items={[__("Documents"), __("Edit document")]} />}
    >
      <form onSubmit={onSubmit}>
        <DialogContent>
          <Textarea
            id="content"
            variant="ghost"
            autogrow
            required
            placeholder={__("Add description")}
            aria-label={__("Description")}
            className="p-6"
            {...register("content")}
          />
        </DialogContent>
        <DialogFooter>
          <Button disabled={isLoading} type="submit">
            {isLoading && <Spinner />}
            {__("Update document")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
