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
  useToast,
} from "@probo/ui";
import { type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CreateDocumentDialogMutation } from "#/__generated__/core/CreateDocumentDialogMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { DocumentClassificationOptions } from "#/components/form/DocumentClassificationOptions";
import { DocumentTypeOptions } from "#/components/form/DocumentTypeOptions";
import { PeopleMultiSelectField } from "#/components/form/PeopleMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

type CreateDocumentDialogProps = {
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
export function CreateDocumentDialog({ trigger, connection }: CreateDocumentDialogProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { toast } = useToast();
  const documentSchema = z.object({
    title: z.string().min(1, t("createDocumentDialog.validation.titleRequired")),
    documentType: z.enum(["OTHER", "GOVERNANCE", "POLICY", "PROCEDURE", "PLAN", "REGISTER", "RECORD", "REPORT", "TEMPLATE"]),
    classification: z.enum(["PUBLIC", "INTERNAL", "CONFIDENTIAL", "SECRET"]),
    defaultApproverIds: z.array(z.string()),
  });

  const { control, handleSubmit, register, formState, reset } = useFormWithSchema(
    documentSchema,
    {
      defaultValues: {
        documentType: "POLICY",
        classification: "INTERNAL",
        defaultApproverIds: [],
      },
    },
  );
  const errors = formState.errors ?? {};
  const [createDocument, isLoading]
    = useMutation<CreateDocumentDialogMutation>(createDocumentMutation);

  const onSubmit = (data: z.infer<typeof documentSchema>) => {
    createDocument({
      variables: {
        input: {
          ...data,
          organizationId,
        },
        connections: [connection],
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: t("createDocumentDialog.errors.title"),
            description: formatError(t("createDocumentDialog.errors.create"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("createDocumentDialog.messages.successTitle"),
          description: t("createDocumentDialog.messages.created"),
          variant: "success",
        });
        dialogRef.current?.close();
        reset();
      },
      onError(error) {
        toast({
          title: t("createDocumentDialog.errors.title"),
          description: error.message,
          variant: "error",
        });
      },
    });
  };

  const dialogRef = useDialogRef();

  return (
    <Dialog
      ref={dialogRef}
      trigger={trigger}
      title={(
        <Breadcrumb
          items={[
            t("createDocumentDialog.breadcrumbs.documents"),
            t("createDocumentDialog.breadcrumbs.new"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent className="grid grid-cols-[1fr_420px]">
          <div className="py-8 px-10 space-y-4">
            <Input
              id="title"
              aria-label={t("createDocumentDialog.fields.title")}
              required
              variant="title"
              placeholder={t("createDocumentDialog.fields.titlePlaceholder")}
              {...register("title")}
            />
          </div>
          {/* Properties form */}
          <div className="py-5 px-6 bg-subtle">
            <Label>{t("createDocumentDialog.properties.title")}</Label>
            <PropertyRow label={t("createDocumentDialog.properties.status")}>
              <Badge variant="neutral" size="md">
                {t("createDocumentDialog.status.draft")}
              </Badge>
            </PropertyRow>

            <PropertyRow
              id="documentType"
              label={t("createDocumentDialog.properties.type")}
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
              label={t("createDocumentDialog.properties.classification")}
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

            <PropertyRow label={t("createDocumentDialog.properties.approvers")}>
              <PeopleMultiSelectField
                name="defaultApproverIds"
                control={control}
                organizationId={organizationId}
                placeholder={t("createDocumentDialog.fields.approversPlaceholder")}
              />
            </PropertyRow>

          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isLoading}>
            {t("createDocumentDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
