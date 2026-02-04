import {
  auditStates,
  fileSize,
  formatDate,
  formatDatetime,
  formatError,
  getAuditStateLabel,
  getAuditStateVariant,
  type GraphQLError,
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
  const { __ } = useTranslate();
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
        title: __("Success"),
        description: __("Audit updated successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(
          __("Failed to update audit"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  });

  const handleDeleteReport = () => {
    if (!auditEntry.report || !auditEntry.id) return;

    confirm(
      async () => {
        await deleteAuditReport({ auditId: auditEntry.id! });
      },
      {
        message: sprintf(
          __(
            "This will permanently delete the audit report \"%s\". This action cannot be undone.",
          ),
          auditEntry.report.filename,
        ),
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
            label: __("Audits"),
            to: `/organizations/${organizationId}/audits`,
          },
          {
            label:
              (auditEntry.name || auditEntry.framework?.name)
              ?? __("Unknown Audit"),
          },
        ]}
      />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-3">
            <FrameworkLogo
              name={auditEntry.framework?.name || ""}
              lightLogoURL={auditEntry.framework?.lightLogoURL}
              darkLogoURL={auditEntry.framework?.darkLogoURL}
            />
            <div className="text-2xl">{auditEntry.framework?.name}</div>
          </div>
          <Badge
            variant={getAuditStateVariant(auditEntry.state || "NOT_STARTED")}
          >
            {getAuditStateLabel(__, auditEntry.state || "NOT_STARTED")}
          </Badge>
        </div>
        <ActionDropdown variant="secondary">
          {auditEntry.canDelete && (
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={deleteAudit}
            >
              {__("Delete")}
            </DropdownItem>
          )}
        </ActionDropdown>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <form onSubmit={e => void onSubmit(e)} className="space-y-6">
          <Field label={__("Name")}>
            <Input {...register("name")} placeholder={__("Audit name")} />
          </Field>

          <ControlledField
            control={control}
            name="state"
            type="select"
            label={__("State")}
          >
            {auditStates.map(state => (
              <Option key={state} value={state}>
                {getAuditStateLabel(__, state)}
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
            {formState.isDirty && auditEntry.canUpdate && (
              <Button type="submit" disabled={formState.isSubmitting}>
                {formState.isSubmitting ? __("Updating...") : __("Update")}
              </Button>
            )}
          </div>
        </form>

        <Card padded className="mt-6">
          <div className="space-y-4">
            <h3 className="text-lg font-medium">{__("Audit Report")}</h3>

            {auditEntry.report
              ? (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between p-4 bg-success-50 border border-success-200 rounded-lg">
                      <div className="flex items-center gap-3">
                        <IconArrowInbox className="text-success-600" size={20} />
                        <div className="flex-1">
                          <p className="font-medium text-success-900">
                            {auditEntry.report.filename}
                          </p>
                          <div className="flex items-center gap-4 text-sm text-success-700">
                            <span>{fileSize(__, auditEntry.report.size)}</span>
                            <span>
                              {__("Uploaded")}
                              {" "}
                              {formatDate(auditEntry.report.createdAt)}
                            </span>
                          </div>
                        </div>
                      </div>
                      <ActionDropdown>
                        <DropdownItem
                          onClick={() => {
                            if (auditEntry.report?.downloadUrl) {
                              window.open(auditEntry.report.downloadUrl, "_blank");
                            }
                          }}
                          icon={IconArrowInbox}
                        >
                          {__("Download")}
                        </DropdownItem>
                        <DropdownItem
                          variant="danger"
                          icon={IconTrashCan}
                          onClick={handleDeleteReport}
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
                        "Upload the final audit report document (PDF recommended)",
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
