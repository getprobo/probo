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

import { formatError } from "@probo/helpers";
import { Button, IconCheckmark1, IconCrossLargeX, IconPencil, Input } from "@probo/ui";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentTitleFormMutation } from "#/__generated__/core/DocumentTitleFormMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { z } from "#/lib/zod";

const updateDocumentTitleMutation = graphql`
  mutation DocumentTitleFormMutation($input: UpdateDocumentInput!) {
    updateDocument(input: $input) {
      document {
        id
        title
      }
    }
  }
`;

export function DocumentTitleForm(props: {
  title: string;
  canUpdate: boolean;
  documentId: string;
  documentStatus: string;
  isEditable: boolean;
  onDocumentUpdated: () => void;
}) {
  const { title, canUpdate, documentId, documentStatus, isEditable, onDocumentUpdated } = props;

  const { t } = useTranslation();
  const toast = useToast();
  const schema = z.object({
    title: z.string().min(1, t("documentTitleForm.validation.titleRequired")).max(255),
  });

  const [updateDocument, isUpdating]
    = useMutation<DocumentTitleFormMutation>(updateDocumentTitleMutation);

  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const { register, handleSubmit, reset } = useFormWithSchema(
    schema,
    {
      values: {
        title,
      },
    },
  );

  const canEdit = canUpdate && isEditable && documentStatus !== "ARCHIVED";

  const handleUpdateTitle = (data: { title: string }) => {
    updateDocument({
      variables: {
        input: {
          id: documentId,
          title: data.title,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast.add({
            title: t("documentTitleForm.errors.title"),
            description: formatError(t("documentTitleForm.errors.update"), errors),
            type: "error",
          });
          return;
        }
        setIsEditingTitle(false);
        onDocumentUpdated();
      },
      onError(error) {
        toast.add({
          title: t("documentTitleForm.errors.title"),
          description: error.message,
          type: "error",
        });
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
                reset({ title });
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
              reset({ title });
            }}
          />
        </div>
      )
    : (
        <div className="flex items-center gap-2">
          <span>{title}</span>
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
