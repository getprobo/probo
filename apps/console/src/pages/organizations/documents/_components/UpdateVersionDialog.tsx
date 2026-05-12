import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Spinner,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type RefObject, useCallback, useEffect, useState } from "react";
import { useFragment, useMutation } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { UpdateVersionDialogCreateMutation } from "#/__generated__/core/UpdateVersionDialogCreateMutation.graphql";
import type { UpdateVersionDialogFragment$key } from "#/__generated__/core/UpdateVersionDialogFragment.graphql";
import type { UpdateVersionDialogUpdateMutation } from "#/__generated__/core/UpdateVersionDialogUpdateMutation.graphql";
import { RichTextEditor } from "#/components/form/RichTextEditor";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
  fragment UpdateVersionDialogFragment on Document {
    id
    versions(first: 20) @connection(key: "DocumentLayout_versions") {
      __id
      edges {
        node {
          id
          status
          content
        }
      }
    }
  }
`;

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
          major
          minor
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

const updateDocumentMutation = graphql`
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

type UpdateVersionDialogProps = {
  fKey: UpdateVersionDialogFragment$key;
  ref: RefObject<{ open: () => void } | null>;
};

const versionSchema = z.object({
  content: z.string(),
});

export default function UpdateVersionDialog(props: UpdateVersionDialogProps) {
  const { fKey, ref } = props;

  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const [editorKey, setEditorKey] = useState(0);

  const document = useFragment<UpdateVersionDialogFragment$key>(fragment, fKey);
  const version = document.versions.edges[0].node;
  const isDraft = version.status === "DRAFT";
  const [createDraftDocumentVersion, isCreatingDraft]
    = useMutation<UpdateVersionDialogCreateMutation>(createDraftDocument);
  const [updateDocumentVersion, isUpdating]
    = useMutationWithToasts<UpdateVersionDialogUpdateMutation>(
      updateDocumentMutation,
      {
        successMessage: __("Document updated successfully."),
        errorMessage: __("Failed to update document"),
      },
    );
  const { handleSubmit, setValue, watch } = useFormWithSchema(versionSchema, {
    defaultValues: {
      content: version.content,
    },
  });

  const contentValue = watch("content");

  const onContentChange = useCallback(
    (html: string) => {
      setValue("content", html, { shouldValidate: true });
    },
    [setValue],
  );

  useEffect(() => {
    if (!ref.current) {
      ref.current = {
        open: () => {
          setValue("content", version.content);
          setEditorKey(k => k + 1);
          dialogRef.current?.open();
        },
      };
    }
  });

  if (!version) {
    return;
  }

  const onSubmit = async (data: z.infer<typeof versionSchema>) => {
    if (isDraft) {
      await updateDocumentVersion({
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
          connections: [document.versions.__id],
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

          const newVersionId
            = createResponse?.createDraftDocumentVersion?.documentVersionEdge
              ?.node?.id;
          if (newVersionId && data.content !== version.content) {
            void updateDocumentVersion({
              variables: {
                input: {
                  documentVersionId: newVersionId,
                  content: data.content,
                },
              },
              onSuccess: () => {
                dialogRef.current?.close();
                void navigate(`/organizations/${organizationId}/documents/${document.id}/versions/${newVersionId}`);
              },
            });
          } else {
            dialogRef.current?.close();
            void navigate(`/organizations/${organizationId}/documents/${document.id}/versions/${newVersionId}`);
          }
        },
      });
    }
  };

  const isLoading = isCreatingDraft || isUpdating;

  return (
    <Dialog
      ref={dialogRef}
      title={
        (
          <Breadcrumb
            items={[
              __("Documents"),
              isDraft ? __("Edit draft") : __("Create new draft"),
            ]}
          />
        )
      }
      className="w-[85vw]! max-w-[85vw]! h-[85vh]!"
    >
      <form
        onSubmit={e => void handleSubmit(onSubmit)(e)}
        className="flex flex-col h-[calc(85vh-110px)]"
      >
        <DialogContent className="flex-1 max-h-none! p-0!">
          <div className="flex flex-col h-full overflow-hidden">
            <div className="flex-1 overflow-hidden">
              <RichTextEditor
                key={editorKey}
                value={contentValue}
                onChange={onContentChange}
                placeholder={__("Add content")}
              />
            </div>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button disabled={isLoading} type="submit">
            {isLoading && <Spinner />}
            {isDraft ? __("Update document") : __("Create draft")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
