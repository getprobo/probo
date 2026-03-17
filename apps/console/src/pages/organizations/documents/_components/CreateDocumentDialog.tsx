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
import { type ReactNode } from "react";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CreateDocumentDialogMutation } from "#/__generated__/core/CreateDocumentDialogMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { DocumentClassificationOptions } from "#/components/form/DocumentClassificationOptions";
import { DocumentTypeOptions } from "#/components/form/DocumentTypeOptions";
import { PeopleMultiSelectField } from "#/components/form/PeopleMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
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

const documentSchema = z.object({
  title: z.string().min(1, "Title is required"),
  approverIds: z.array(z.string()).min(1, "At least one approver is required"),
  documentType: z.enum(["OTHER", "ISMS", "POLICY", "PROCEDURE"]),
  classification: z.enum(["PUBLIC", "INTERNAL", "CONFIDENTIAL", "SECRET"]),
});

/**
 * Dialog to create or update a document
 */
export function CreateDocumentDialog({ trigger, connection }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const { control, handleSubmit, register, formState, reset } = useFormWithSchema(
    documentSchema,
    {
      defaultValues: {
        documentType: "POLICY",
        classification: "INTERNAL",
      },
    },
  );
  const errors = formState.errors ?? {};
  const [createDocument, isLoading]
    = useMutationWithToasts<CreateDocumentDialogMutation>(createDocumentMutation);

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
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent className="grid grid-cols-[1fr_420px]">
          <div className="py-8 px-10 space-y-4">
            <Input
              id="title"
              aria-label={__("Title")}
              required
              variant="title"
              placeholder={__("Document title")}
              {...register("title")}
            />
          </div>
          {/* Properties form */}
          <div className="py-5 px-6 bg-subtle">
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

            <PropertyRow
              id="approverIds"
              label={__("Approvers")}
              error={errors.approverIds?.message}
            >
              <PeopleMultiSelectField
                name="approverIds"
                control={control}
                organizationId={organizationId}
                placeholder={__("Add approvers...")}
              />
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
