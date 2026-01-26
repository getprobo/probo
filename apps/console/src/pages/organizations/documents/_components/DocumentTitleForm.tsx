import { useTranslate } from "@probo/i18n";
import { Button, IconCheckmark1, IconCrossLargeX, IconPencil, Input } from "@probo/ui";
import { useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { DocumentTitleFormFragment$key } from "#/__generated__/core/DocumentTitleFormFragment.graphql";
import type { DocumentTitleFormMutation } from "#/__generated__/core/DocumentTitleFormMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const updateDocumentTitleMutation = graphql`
  mutation DocumentTitleFormMutation($input: UpdateDocumentInput!) {
    updateDocument(input: $input) {
      document {
        ...DocumentTitleFormFragment
      }
    }
  }
`;

const fragment = graphql`
  fragment DocumentTitleFormFragment on Document {
    id
    title
    canUpdate: permission(action: "core:document:update")
  }
`;

const schema = z.object({
  title: z.string().min(1, "Title is required").max(255),
});

export function DocumentTitleForm(props: { fKey: DocumentTitleFormFragment$key }) {
  const { fKey } = props;

  const { __ } = useTranslate();

  const document = useFragment<DocumentTitleFormFragment$key>(fragment, fKey);
  const [updateDocument, isUpdatingDocument]
    = useMutationWithToasts<DocumentTitleFormMutation>(
      updateDocumentTitleMutation,
      {
        successMessage: __("Document updated successfully."),
        errorMessage: __("Failed to update document"),
      },
    );

  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const { register, handleSubmit, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        title: document.title,
      },
    },
  );

  const handleUpdateTitle = async (data: { title: string }) => {
    await updateDocument({
      variables: {
        input: {
          id: document.id,
          title: data.title,
        },
      },
      onSuccess: () => {
        setIsEditingTitle(false);
      },
    });
  };

  return isEditingTitle
    ? (
        <div className="flex items-center gap-2">
          <Input
            {...register("title")}
            variant="title"
            className="flex-1"
            autoFocus
            onKeyDown={(e) => {
              if (e.key === "Escape") {
                setIsEditingTitle(false);
                reset();
              }
              if (e.key === "Enter") {
                void handleSubmit(handleUpdateTitle)();
              }
            }}
          />
          <Button
            variant="quaternary"
            icon={IconCheckmark1}
            onClick={() => void handleSubmit(handleUpdateTitle)()}
            disabled={isUpdatingDocument}
          />
          <Button
            variant="quaternary"
            icon={IconCrossLargeX}
            onClick={() => {
              setIsEditingTitle(false);
              reset();
            }}
          />
        </div>
      )
    : (
        <div className="flex items-center gap-2">
          <span>{document.title}</span>
          {document.canUpdate && (
            <Button
              variant="quaternary"
              icon={IconPencil}
              onClick={() => setIsEditingTitle(true)}
            />
          )}
        </div>
      );
}
