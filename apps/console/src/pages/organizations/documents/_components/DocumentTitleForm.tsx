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

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, IconCheckmark1, IconCrossLargeX, IconPencil, Input, useToast } from "@probo/ui";
import { useState } from "react";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { DocumentTitleFormFragment$key } from "#/__generated__/core/DocumentTitleFormFragment.graphql";
import type { DocumentTitleFormMutation } from "#/__generated__/core/DocumentTitleFormMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const updateDocumentTitleMutation = graphql`
  mutation DocumentTitleFormMutation($input: UpdateDocumentInput!) {
    updateDocument(input: $input) {
      documentVersion {
        ...DocumentTitleFormFragment
      }
    }
  }
`;

const fragment = graphql`
  fragment DocumentTitleFormFragment on DocumentVersion {
    title
    status
    canUpdate: permission(action: "core:document:update")
  }
`;

const schema = z.object({
  title: z.string().min(1, "Title is required").max(255),
});

export function DocumentTitleForm(props: {
  fKey: DocumentTitleFormFragment$key;
  documentId: string;
  documentStatus: string;
  isEditable: boolean;
  onDocumentUpdated: () => void;
}) {
  const { fKey, documentId, documentStatus, isEditable, onDocumentUpdated } = props;

  const { __ } = useTranslate();
  const { toast } = useToast();

  const version = useFragment<DocumentTitleFormFragment$key>(fragment, fKey);
  const [updateDocument, isUpdating]
    = useMutation<DocumentTitleFormMutation>(updateDocumentTitleMutation);

  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const { register, handleSubmit, reset } = useFormWithSchema(
    schema,
    {
      values: {
        title: version.title,
      },
    },
  );

  const isDraft = version.status === "DRAFT";
  const canEdit = version.canUpdate && isEditable && documentStatus !== "ARCHIVED";

  const handleUpdateTitle = (data: { title: string }) => {
    updateDocument({
      variables: {
        input: {
          id: documentId,
          title: data.title,
        },
      },
      onCompleted(data, errors) {
        if (errors?.length) {
          toast({ title: __("Error"), description: formatError(__("Failed to update document"), errors), variant: "error" });
          return;
        }
        setIsEditingTitle(false);
        const draftReturned = !!data.updateDocument.documentVersion;
        if (isDraft !== draftReturned) {
          onDocumentUpdated();
        }
      },
      onError(error) {
        toast({ title: __("Error"), description: error.message, variant: "error" });
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
                reset({ title: version.title });
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
            disabled={isUpdating}
          />
          <Button
            variant="quaternary"
            icon={IconCrossLargeX}
            onClick={() => {
              setIsEditingTitle(false);
              reset({ title: version.title });
            }}
          />
        </div>
      )
    : (
        <div className="flex items-center gap-2">
          <span>{version.title}</span>
          {canEdit && (
            <Button
              variant="quaternary"
              icon={IconPencil}
              onClick={() => setIsEditingTitle(true)}
            />
          )}
        </div>
      );
}
