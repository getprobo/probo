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
import { type ReactNode, useCallback } from "react";
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
          canSendSigningNotifications: permission(action: "core:document:send-signing-notifications")
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

  const { control, handleSubmit, register, formState, reset, setValue }
    = useDocumentForm();
  const errors = formState.errors ?? {};
  const [createDocument, isLoading]
    = useMutationWithToasts<CreateDocumentDialogMutation>(createDocumentMutation);

  const onContentChange = useCallback(
    (html: string) => {
      setValue("content", html, { shouldValidate: true });
    },
    [setValue],
  );

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
                onChange={onContentChange}
                placeholder={__("Add content")}
              />
            </div>
          </div>
          {/* Properties form */}
          <div className="py-5 px-6 bg-subtle overflow-y-auto">
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
