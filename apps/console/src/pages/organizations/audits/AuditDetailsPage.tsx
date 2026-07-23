// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import {
  auditStates,
  formatDatetime,
  formatError,
  getAuditStateVariant,
  type GraphQLError,
} from "@probo/helpers";
import { dateFormat, fileSize } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  Card,
  DropdownItem,
  Dropzone,
  Field,
  FrameworkLogo,
  IconArrowInbox,
  IconTrashCan,
  Input,
  Option,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate } from "react-router";
import { z } from "zod";

import type { AuditGraphNodeQuery } from "#/__generated__/core/AuditGraphNodeQuery.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  auditNodeQuery,
  useDeleteAudit,
  useDeleteAuditReport,
  useUpdateAudit,
  useUploadAuditReport,
} from "../../../hooks/graph/AuditGraph";

const updateAuditSchema = z.object({
  name: z.string().nullable().optional(),
  validFrom: z.string().optional(),
  validUntil: z.string().optional(),
  state: z.enum([
    "NOT_STARTED",
    "IN_PROGRESS",
    "COMPLETED",
    "REJECTED",
    "OUTDATED",
  ]),
});

type Props = {
  queryRef: PreloadedQuery<AuditGraphNodeQuery>;
};

export default function AuditDetailsPage(props: Props) {
  const audit = usePreloadedQuery<AuditGraphNodeQuery>(
    auditNodeQuery,
    props.queryRef,
  );
  const auditEntry = audit.node;
  const { i18n, t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const deleteAudit = useDeleteAudit(
    { id: auditEntry.id!, framework: { name: auditEntry.framework!.name } },
    ConnectionHandler.getConnectionID(organizationId, "AuditsPage_audits"),
    () => void navigate(`/organizations/${organizationId}/audits`),
  );

  const { control, formState, handleSubmit, register, reset }
    = useFormWithSchema(updateAuditSchema, {
      defaultValues: {
        name: auditEntry.name || null,
        validFrom: auditEntry.validFrom?.split("T")[0] || "",
        validUntil: auditEntry.validUntil?.split("T")[0] || "",
        state: auditEntry.state || "NOT_STARTED",
      },
    });

  const updateAudit = useUpdateAudit();
  const [uploadAuditReport, isUploading] = useUploadAuditReport();
  const deleteAuditReport = useDeleteAuditReport();
  const confirm = useConfirm();
  const { toast } = useToast();

  const onSubmit = handleSubmit(async (formData) => {
    if (!auditEntry.id) return;

    try {
      await updateAudit({
        id: auditEntry.id,
        name: formData.name || null,
        validFrom: formatDatetime(formData.validFrom) ?? null,
        validUntil: formatDatetime(formData.validUntil) ?? null,
        state: formData.state,
      });
      reset(formData);
      toast({
        title: t("auditDetailsPage.messages.success"),
        description: t("auditDetailsPage.messages.updated"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: t("auditDetailsPage.messages.error"),
        description: formatError(
          t("auditDetailsPage.errors.update"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  });

  const handleDeleteReport = () => {
    if (!auditEntry.reportFile || !auditEntry.id) return;

    confirm(
      async () => {
        await deleteAuditReport({ auditId: auditEntry.id! });
      },
      {
        message: t("auditDetailsPage.deleteReportConfirmation", {
          name: auditEntry.reportFile.fileName,
        }),
      },
    );
  };

  const handleUploadFile = async (files: File[]) => {
    if (files.length > 0 && auditEntry.id) {
      await uploadAuditReport({
        auditId: auditEntry.id,
        file: files[0],
      });
      window.location.reload();
    }
  };

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: t("auditDetailsPage.breadcrumb.audits"),
            to: `/organizations/${organizationId}/audits`,
          },
          {
            label:
              (auditEntry.name || auditEntry.framework?.name)
              ?? t("auditDetailsPage.unknownAudit"),
          },
        ]}
      />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-3">
            <FrameworkLogo
              name={auditEntry.framework?.name || ""}
              lightLogoURL={auditEntry.framework?.lightLogo?.downloadUrl}
              darkLogoURL={auditEntry.framework?.darkLogo?.downloadUrl}
            />
            <div className="text-2xl">{auditEntry.framework?.name}</div>
          </div>
          <Badge
            variant={getAuditStateVariant(auditEntry.state || "NOT_STARTED")}
          >
            {t(`auditDetailsPage.states.${(auditEntry.state || "NOT_STARTED").toLowerCase()}`)}
          </Badge>
        </div>
        <ActionDropdown variant="secondary">
          {auditEntry.canDelete && (
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={deleteAudit}
            >
              {t("auditDetailsPage.actions.delete")}
            </DropdownItem>
          )}
        </ActionDropdown>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <form onSubmit={e => void onSubmit(e)} className="space-y-6">
          <Field label={t("auditDetailsPage.fields.name")}>
            <Input
              {...register("name")}
              placeholder={t("auditDetailsPage.fields.namePlaceholder")}
            />
          </Field>

          <ControlledField
            control={control}
            name="state"
            type="select"
            label={t("auditDetailsPage.fields.state")}
          >
            {auditStates.map(state => (
              <Option key={state} value={state}>
                {t(`auditDetailsPage.states.${state.toLowerCase()}`)}
              </Option>
            ))}
          </ControlledField>

          <Field label={t("auditDetailsPage.fields.validFrom")}>
            <Input {...register("validFrom")} type="date" />
          </Field>

          <Field label={t("auditDetailsPage.fields.validUntil")}>
            <Input {...register("validUntil")} type="date" />
          </Field>

          <div className="flex justify-end">
            {formState.isDirty && auditEntry.canUpdate && (
              <Button type="submit" disabled={formState.isSubmitting}>
                {formState.isSubmitting
                  ? t("auditDetailsPage.actions.updating")
                  : t("auditDetailsPage.actions.update")}
              </Button>
            )}
          </div>
        </form>

        <Card padded className="mt-6">
          <div className="space-y-4">
            <h3 className="text-lg font-medium">
              {t("auditDetailsPage.report.title")}
            </h3>

            {auditEntry.reportFile
              ? (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between p-4 bg-success-50 border border-success-200 rounded-lg">
                      <div className="flex items-center gap-3">
                        <IconArrowInbox className="text-success-600" size={20} />
                        <div className="flex-1">
                          <p className="font-medium text-success-900">
                            {auditEntry.reportFile.fileName}
                          </p>
                          <div className="flex items-center gap-4 text-sm text-success-700">
                            <span>
                              {fileSize(auditEntry.reportFile.size, t)}
                            </span>
                            <span>
                              {t("auditDetailsPage.report.uploaded", {
                                date: dateFormat(
                                  i18n.language,
                                  auditEntry.reportFile.createdAt,
                                ),
                              })}
                            </span>
                          </div>
                        </div>
                      </div>
                      <ActionDropdown>
                        <DropdownItem
                          onClick={() => {
                            if (auditEntry.reportFile?.downloadUrl) {
                              window.open(auditEntry.reportFile.downloadUrl, "_blank", "noopener,noreferrer");
                            }
                          }}
                          icon={IconArrowInbox}
                        >
                          {t("auditDetailsPage.actions.download")}
                        </DropdownItem>
                        <DropdownItem
                          variant="danger"
                          icon={IconTrashCan}
                          onClick={handleDeleteReport}
                        >
                          {t("auditDetailsPage.actions.delete")}
                        </DropdownItem>
                      </ActionDropdown>
                    </div>
                  </div>
                )
              : (
                  <div className="space-y-4">
                    <p className="text-neutral-600">
                      {t("auditDetailsPage.report.uploadDescription")}
                    </p>
                    <Dropzone
                      description={t("auditDetailsPage.report.dropzoneDescription")}
                      isUploading={isUploading}
                      onDrop={files => void handleUploadFile(files)}
                      accept={{ "application/pdf": [".pdf"] }}
                      maxSize={25}
                    />
                  </div>
                )}
          </div>
        </Card>
      </div>
    </div>
  );
}
