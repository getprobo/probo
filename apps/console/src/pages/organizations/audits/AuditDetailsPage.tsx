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
  toPeriod,
} from "@probo/helpers";
import { dateFormat, fileSize } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
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
  Select,
  useConfirm,
  useToast,
} from "@probo/ui";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  type PreloadedQuery,
  useLazyLoadQuery,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { AuditDetailsPageFrameworksQuery } from "#/__generated__/core/AuditDetailsPageFrameworksQuery.graphql";
import type { AuditGraphNodeQuery } from "#/__generated__/core/AuditGraphNodeQuery.graphql";
import { auditFrameworkNames } from "#/components/audits/auditFrameworkNames";
import { ControlledField } from "#/components/form/ControlledField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { z } from "#/lib/zod";

import {
  auditNodeQuery,
  useDeleteAudit,
  useDeleteAuditReport,
  useLinkAuditFramework,
  useUnlinkAuditFramework,
  useUpdateAudit,
  useUploadAuditReport,
} from "../../../hooks/graph/AuditGraph";

const organizationFrameworksQuery = graphql`
  query AuditDetailsPageFrameworksQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        frameworks(first: 100) {
          edges {
            node {
              id
              name
            }
          }
        }
      }
    }
  }
`;

const updateAuditSchema = z.object({
  name: z.string().nullable().optional(),
  firm: z.string().nullable().optional(),
  validFrom: z.string().optional(),
  validUntil: z.string().optional(),
  auditStartDate: z.string().optional(),
  auditEndDate: z.string().optional(),
  state: z.enum(auditStates),
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
  const frameworks = auditEntry.frameworks?.edges?.map(edge => edge.node) ?? [];
  const frameworkLabel = auditFrameworkNames(auditEntry.frameworks);
  const frameworksConnectionId = auditEntry.frameworks?.__id ?? "";
  const linkFramework = useLinkAuditFramework(frameworksConnectionId);
  const unlinkFramework = useUnlinkAuditFramework(frameworksConnectionId);

  const deleteAudit = useDeleteAudit(
    { id: auditEntry.id!, frameworks: auditEntry.frameworks },
    ConnectionHandler.getConnectionID(organizationId, "AuditsPage_audits"),
    () => void navigate(`/organizations/${organizationId}/governance/audits`),
  );

  const { control, formState, handleSubmit, register, reset }
    = useFormWithSchema(updateAuditSchema, {
      defaultValues: {
        name: auditEntry.name || null,
        firm: auditEntry.firm || null,
        validFrom: auditEntry.validity?.start?.split("T")[0] || "",
        validUntil: auditEntry.validity?.end?.split("T")[0] || "",
        auditStartDate: auditEntry.auditDates?.start?.split("T")[0] || "",
        auditEndDate: auditEntry.auditDates?.end?.split("T")[0] || "",
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
        firm: formData.firm || null,
        validity: toPeriod(
          formatDatetime(formData.validFrom),
          formatDatetime(formData.validUntil),
        ),
        auditDates: toPeriod(
          formatDatetime(formData.auditStartDate),
          formatDatetime(formData.auditEndDate),
        ),
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
      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-3">
            {frameworks.map(framework => (
              <FrameworkLogo
                key={framework.id}
                name={framework.name}
                lightLogoURL={framework.lightLogo?.downloadUrl}
                darkLogoURL={framework.darkLogo?.downloadUrl}
              />
            ))}
            <div className="text-2xl">
              {frameworkLabel || t("auditsPage.row.unknownFramework")}
            </div>
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

      <Card padded className="space-y-4">
        <h3 className="text-lg font-medium">
          {t("auditDetailsPage.frameworks.title")}
        </h3>
        {frameworks.length === 0 && (
          <p className="text-txt-secondary">
            {t("auditDetailsPage.frameworks.empty")}
          </p>
        )}
        <ul className="space-y-2">
          {frameworks.map(framework => (
            <li
              key={framework.id}
              className="flex items-center justify-between gap-3"
            >
              <div className="flex items-center gap-3">
                <FrameworkLogo
                  name={framework.name}
                  lightLogoURL={framework.lightLogo?.downloadUrl}
                  darkLogoURL={framework.darkLogo?.downloadUrl}
                />
                <span>{framework.name}</span>
              </div>
              {auditEntry.canUpdate && (
                <Button
                  variant="secondary"
                  icon={IconTrashCan}
                  onClick={() => {
                    if (!auditEntry.id) {
                      return;
                    }
                    void unlinkFramework(auditEntry.id, framework.id);
                  }}
                >
                  {t("auditDetailsPage.frameworks.unlink")}
                </Button>
              )}
            </li>
          ))}
        </ul>
        {auditEntry.canUpdate && auditEntry.id && (
          <Suspense fallback={null}>
            <LinkFrameworkSelect
              organizationId={organizationId}
              linkedIds={frameworks.map(framework => framework.id)}
              onLink={frameworkId => void linkFramework(auditEntry.id!, frameworkId)}
            />
          </Suspense>
        )}
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <form onSubmit={e => void onSubmit(e)} className="space-y-6">
          <Field label={t("auditDetailsPage.fields.name")}>
            <Input
              {...register("name")}
              placeholder={t("auditDetailsPage.fields.namePlaceholder")}
            />
          </Field>

          <Field label={t("auditDetailsPage.fields.firm")}>
            <Input
              {...register("firm")}
              placeholder={t("auditDetailsPage.fields.firmPlaceholder")}
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

          <Field label={t("auditDetailsPage.fields.auditStartDate")}>
            <Input {...register("auditStartDate")} type="date" />
          </Field>

          <Field label={t("auditDetailsPage.fields.auditEndDate")}>
            <Input {...register("auditEndDate")} type="date" />
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

function LinkFrameworkSelect({
  organizationId,
  linkedIds,
  onLink,
}: {
  organizationId: string;
  linkedIds: string[];
  onLink: (frameworkId: string) => void;
}) {
  const { t } = useTranslation();
  const data = useLazyLoadQuery<AuditDetailsPageFrameworksQuery>(
    organizationFrameworksQuery,
    { organizationId },
  );
  const linked = new Set(linkedIds);
  const frameworks = data.organization?.frameworks?.edges
    ?.map(edge => edge.node)
    .filter((node): node is NonNullable<typeof node> => node != null)
    .filter(framework => !linked.has(framework.id)) ?? [];

  if (frameworks.length === 0) {
    return null;
  }

  return (
    <Select
      variant="editor"
      placeholder={t("auditDetailsPage.frameworks.addPlaceholder")}
      onValueChange={onLink}
      value=""
    >
      {frameworks.map(framework => (
        <Option key={framework.id} value={framework.id}>
          {framework.name}
        </Option>
      ))}
    </Select>
  );
}
