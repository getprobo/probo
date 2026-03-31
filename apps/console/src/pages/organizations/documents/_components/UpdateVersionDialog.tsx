// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
import { type RefObject, useEffect } from "react";
import { useFragment, useMutation } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { UpdateVersionDialogCreateMutation } from "#/__generated__/core/UpdateVersionDialogCreateMutation.graphql";
import type { UpdateVersionDialogFragment$key } from "#/__generated__/core/UpdateVersionDialogFragment.graphql";
import type { UpdateVersionDialogUpdateMutation } from "#/__generated__/core/UpdateVersionDialogUpdateMutation.graphql";
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
  const { handleSubmit, register } = useFormWithSchema(versionSchema, {
    defaultValues: {
      content: version.content,
    },
  });

  useEffect(() => {
    if (!ref.current) {
      ref.current = {
        open: () => {
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
      title={<Breadcrumb items={[__("Documents"), __("Edit document")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent>
          <Textarea
            id="content"
            variant="ghost"
            autogrow
            required
            placeholder={__("Add content")}
            aria-label={__("Content")}
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
