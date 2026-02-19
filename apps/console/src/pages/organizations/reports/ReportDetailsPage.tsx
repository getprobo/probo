import {
  fileSize,
  formatDate,
  formatDatetime,
  formatError,
  getReportStateLabel,
  getReportStateVariant,
  type GraphQLError,
  reportStates,
  sprintf,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
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
import {
  ConnectionHandler,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { ReportDetailsPageQuery } from "#/__generated__/core/ReportDetailsPageQuery.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  useDeleteReport,
  useDeleteReportFile,
  useUpdateReport,
  useUploadReportFile,
} from "../../../hooks/graph/ReportGraph";

export const reportDetailsPageQuery = graphql`
  query ReportDetailsPageQuery($reportId: ID!) {
    node(id: $reportId) {
      ... on Report {
        id
        name
        frameworkType
        validFrom
        validUntil
        file {
          id
          fileName
          mimeType
          size
          downloadUrl
          createdAt
        }
        reportUrl
        state
        framework {
          id
          name
          lightLogoURL
          darkLogoURL
        }
        organization {
          id
          name
        }
        createdAt
        updatedAt
        canUpdate: permission(action: "core:report:update")
        canDelete: permission(action: "core:report:delete")
      }
    }
  }
`;

const updateReportSchema = z.object({
  name: z.string().nullable().optional(),
  frameworkType: z.string().nullable().optional(),
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
  queryRef: PreloadedQuery<ReportDetailsPageQuery>;
};

export default function ReportDetailsPage(props: Props) {
  const report = usePreloadedQuery<ReportDetailsPageQuery>(
    reportDetailsPageQuery,
    props.queryRef,
  );
  const reportEntry = report.node;
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const deleteReport = useDeleteReport(
    { id: reportEntry.id!, framework: { name: reportEntry.framework!.name } },
    ConnectionHandler.getConnectionID(organizationId, "ReportsPage_reports"),
    () => void navigate(`/organizations/${organizationId}/reports`),
  );

  const { control, formState, handleSubmit, register, reset }
    = useFormWithSchema(updateReportSchema, {
      defaultValues: {
        name: reportEntry.name || null,
        frameworkType: reportEntry.frameworkType || null,
        validFrom: reportEntry.validFrom?.split("T")[0] || "",
        validUntil: reportEntry.validUntil?.split("T")[0] || "",
        state: reportEntry.state || "NOT_STARTED",
      },
    });

  const updateReport = useUpdateReport();
  const [uploadReportFile, isUploading] = useUploadReportFile();
  const deleteReportFile = useDeleteReportFile();
  const confirm = useConfirm();
  const { toast } = useToast();

  const onSubmit = handleSubmit(async (formData) => {
    if (!reportEntry.id) return;

    try {
      await updateReport({
        id: reportEntry.id,
        name: formData.name || null,
        frameworkType: formData.frameworkType || null,
        validFrom: formatDatetime(formData.validFrom) ?? null,
        validUntil: formatDatetime(formData.validUntil) ?? null,
        state: formData.state,
      });
      reset(formData);
      toast({
        title: __("Success"),
        description: __("Report updated successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(
          __("Failed to update report"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  });

  const handleDeleteFile = () => {
    if (!reportEntry.file || !reportEntry.id) return;

    confirm(
      async () => {
        await deleteReportFile({ reportId: reportEntry.id! });
      },
      {
        message: sprintf(
          __(
            "This will permanently delete the report file \"%s\". This action cannot be undone.",
          ),
          reportEntry.file.fileName,
        ),
      },
    );
  };

  const handleUploadFile = async (files: File[]) => {
    if (files.length > 0 && reportEntry.id) {
      await uploadReportFile({
        reportId: reportEntry.id,
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
            label: __("Reports"),
            to: `/organizations/${organizationId}/reports`,
          },
          {
            label:
              (reportEntry.name || reportEntry.framework?.name)
              ?? __("Unknown Report"),
          },
        ]}
      />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-3">
            <FrameworkLogo
              name={reportEntry.framework?.name || ""}
              lightLogoURL={reportEntry.framework?.lightLogoURL}
              darkLogoURL={reportEntry.framework?.darkLogoURL}
            />
            <div className="text-2xl">{reportEntry.framework?.name}</div>
          </div>
          <Badge
            variant={getReportStateVariant(reportEntry.state || "NOT_STARTED")}
          >
            {getReportStateLabel(__, reportEntry.state || "NOT_STARTED")}
          </Badge>
        </div>
        <ActionDropdown variant="secondary">
          {reportEntry.canDelete && (
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={deleteReport}
            >
              {__("Delete")}
            </DropdownItem>
          )}
        </ActionDropdown>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <form onSubmit={e => void onSubmit(e)} className="space-y-6">
          <Field label={__("Name")}>
            <Input {...register("name")} placeholder={__("Report name")} />
          </Field>

          <Field label={__("Framework Type")}>
            <Input {...register("frameworkType")} placeholder={__("Framework type")} />
          </Field>

          <ControlledField
            control={control}
            name="state"
            type="select"
            label={__("State")}
          >
            {reportStates.map(state => (
              <Option key={state} value={state}>
                {getReportStateLabel(__, state)}
              </Option>
            ))}
          </ControlledField>

          <Field label={__("Valid From")}>
            <Input {...register("validFrom")} type="date" />
          </Field>

          <Field label={__("Valid Until")}>
            <Input {...register("validUntil")} type="date" />
          </Field>

          <div className="flex justify-end">
            {formState.isDirty && reportEntry.canUpdate && (
              <Button type="submit" disabled={formState.isSubmitting}>
                {formState.isSubmitting ? __("Updating...") : __("Update")}
              </Button>
            )}
          </div>
        </form>

        <Card padded className="mt-6">
          <div className="space-y-4">
            <h3 className="text-lg font-medium">{__("Report File")}</h3>

            {reportEntry.file
              ? (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between p-4 bg-success-50 border border-success-200 rounded-lg">
                      <div className="flex items-center gap-3">
                        <IconArrowInbox className="text-success-600" size={20} />
                        <div className="flex-1">
                          <p className="font-medium text-success-900">
                            {reportEntry.file.fileName}
                          </p>
                          <div className="flex items-center gap-4 text-sm text-success-700">
                            <span>{fileSize(__, reportEntry.file.size)}</span>
                            <span>
                              {__("Uploaded")}
                              {" "}
                              {formatDate(reportEntry.file.createdAt)}
                            </span>
                          </div>
                        </div>
                      </div>
                      <ActionDropdown>
                        <DropdownItem
                          onClick={() => {
                            if (reportEntry.file?.downloadUrl) {
                              window.open(reportEntry.file.downloadUrl, "_blank");
                            }
                          }}
                          icon={IconArrowInbox}
                        >
                          {__("Download")}
                        </DropdownItem>
                        <DropdownItem
                          variant="danger"
                          icon={IconTrashCan}
                          onClick={handleDeleteFile}
                        >
                          {__("Delete")}
                        </DropdownItem>
                      </ActionDropdown>
                    </div>
                  </div>
                )
              : (
                  <div className="space-y-4">
                    <p className="text-neutral-600">
                      {__(
                        "Upload the final report document (PDF recommended)",
                      )}
                    </p>
                    <Dropzone
                      description={__(
                        "Only PDF, DOCX files up to 25MB are allowed",
                      )}
                      isUploading={isUploading}
                      onDrop={files => void handleUploadFile(files)}
                      accept={{
                        "application/pdf": [".pdf"],
                        "application/msword": [".doc"],
                        "application/vnd.openxmlformats-officedocument.wordprocessingml.document": [".docx"],
                        "application/vnd.oasis.opendocument.text": [".odt"],
                      }}
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
