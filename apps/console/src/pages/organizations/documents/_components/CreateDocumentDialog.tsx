// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
  Badge,
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Input,
  Label,
  PropertyRow,
  useDialogRef,
} from "@probo/ui";
import { clsx } from "clsx";
import { type ReactNode, useCallback, useState } from "react";
import { graphql } from "relay-runtime";
import type { z } from "zod";

import type { CreateDocumentDialogMutation } from "#/__generated__/core/CreateDocumentDialogMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { DocumentClassificationOptions } from "#/components/form/DocumentClassificationOptions";
import { DocumentTypeOptions } from "#/components/form/DocumentTypeOptions";
import { RichTextEditor } from "#/components/form/RichTextEditor";
import { documentSchema, useDocumentForm } from "#/hooks/forms/useDocumentForm";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { type DocumentTemplate, documentTemplates } from "./templates";

type Props = {
  trigger?: ReactNode;
  connection: string;
};

const createDocumentMutation = graphql`
  mutation CreateDocumentDialogMutation(
    $input: CreateDocumentInput!
    $connections: [ID!]!
  ) {
    createDocument(input: $input) {
      documentEdge @prependEdge(connections: $connections) {
        node {
          id
          canUpdate: permission(action: "core:document:update")
          canDelete: permission(action: "core:document:delete")
          canRequestSignatures: permission(action: "core:document-version:request-signature")
          canArchive: permission(action: "core:document:archive")
          canUnarchive: permission(action: "core:document:unarchive")
          ...DocumentListItemFragment
        }
      }
    }
  }
`;

/**
 * Dialog to create or update a document
 */
export function CreateDocumentDialog({ trigger, connection }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null);
  const [editorKey, setEditorKey] = useState(0);

  const { control, handleSubmit, register, formState, reset, setValue, watch }
    = useDocumentForm();
  const errors = formState.errors ?? {};
  const contentValue = watch("content");
  const [createDocument, isLoading]
    = useMutationWithToasts<CreateDocumentDialogMutation>(createDocumentMutation);

  const onContentChange = useCallback(
    (html: string) => {
      setValue("content", html, { shouldValidate: true });
    },
    [setValue],
  );

  const applyTemplate = (template: DocumentTemplate) => {
    setValue("title", template.title);
    setValue("content", template.content, { shouldValidate: true });
    setValue("documentType", template.documentType);
    setValue("classification", template.classification);
    setSelectedTemplate(template.id);
    setEditorKey(k => k + 1);
  };

  const clearTemplate = () => {
    reset();
    setSelectedTemplate(null);
    setEditorKey(k => k + 1);
  };

  const onSubmit = async (data: z.infer<typeof documentSchema>) => {
    await createDocument({
      variables: {
        input: {
          ...data,
          organizationId,
        },
        connections: [connection],
      },
      successMessage: __("Document created successfully."),
      errorMessage: __("Failed to create document"),
      onSuccess: () => {
        dialogRef.current?.close();
        reset();
        setSelectedTemplate(null);
        setEditorKey(k => k + 1);
      },
    });
  };

  const dialogRef = useDialogRef();

  return (
    <Dialog
      ref={dialogRef}
      trigger={trigger}
      title={<Breadcrumb items={[__("Documents"), __("New Document")]} />}
      className="w-[85vw]! max-w-[85vw]! h-[85vh]!"
      onClose={() => {
        reset();
        setSelectedTemplate(null);
        setEditorKey(k => k + 1);
      }}
    >
      <form
        onSubmit={e => void handleSubmit(onSubmit)(e)}
        className="flex flex-col h-[calc(85vh-110px)]"
      >
        <DialogContent className="grid grid-cols-[1fr_420px] flex-1 max-h-none!">
          <div className="flex flex-col overflow-hidden">
            <div className="py-4 px-10 shrink-0">
              <Input
                id="title"
                aria-label={__("Title")}
                required
                variant="title"
                placeholder={__("Document title")}
                {...register("title")}
              />
            </div>
            <div className="flex-1 overflow-hidden">
              <RichTextEditor
                key={editorKey}
                value={contentValue}
                onChange={onContentChange}
                placeholder={__("Add content")}
              />
            </div>
          </div>
          {/* Properties & templates sidebar */}
          <div className="py-5 px-6 bg-subtle overflow-y-auto">
            <Label>{__("Template")}</Label>
            <div className="flex flex-wrap gap-1.5 mt-2 mb-5">
              <button
                type="button"
                onClick={clearTemplate}
                className={clsx(
                  "flex items-center gap-1.5 px-2.5 py-1 rounded-md border text-xs font-medium transition-all cursor-pointer",
                  selectedTemplate === null
                    ? "border-primary bg-primary/5 text-primary ring-1 ring-primary/20"
                    : "border-border-low bg-level-1 text-txt-secondary hover:border-txt-tertiary hover:text-txt-primary",
                )}
              >
                {__("Blank")}
              </button>
              {documentTemplates.map((template) => {
                const Icon = template.icon;
                const isActive = selectedTemplate === template.id;
                return (
                  <button
                    key={template.id}
                    type="button"
                    onClick={() => applyTemplate(template)}
                    className={clsx(
                      "flex items-center gap-1.5 px-2.5 py-1 rounded-md border text-xs font-medium transition-all cursor-pointer",
                      isActive
                        ? "border-primary bg-primary/5 text-primary ring-1 ring-primary/20"
                        : "border-border-low bg-level-1 text-txt-secondary hover:border-txt-tertiary hover:text-txt-primary",
                    )}
                  >
                    <Icon size={12} />
                    {__(template.name)}
                  </button>
                );
              })}
            </div>

            <Label>{__("Properties")}</Label>
            <PropertyRow label={__("Status")}>
              <Badge variant="neutral" size="md">
                {__("Draft")}
              </Badge>
            </PropertyRow>

            <PropertyRow
              id="documentType"
              label={__("Type")}
              error={errors.documentType?.message}
            >
              <ControlledField
                control={control}
                name="documentType"
                type="select"
              >
                <DocumentTypeOptions />
              </ControlledField>
            </PropertyRow>

            <PropertyRow
              id="classification"
              label={__("Classification")}
              error={errors.classification?.message}
            >
              <ControlledField
                control={control}
                name="classification"
                type="select"
              >
                <DocumentClassificationOptions />
              </ControlledField>
            </PropertyRow>

          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isLoading}>
            {__("Create document")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
