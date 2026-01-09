import { useNavigate, useParams } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import {
  ActionDropdown,
  Breadcrumb,
  DropdownItem,
  IconTrashCan,
  IconArrowDown,
  PageHeader,
  Card,
  Field,
  Button,
  IconPencil,
  IconCheckmark1,
  IconCrossLargeX,
  Input,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import type {
  StateOfApplicabilityGraphNodeQuery,
} from "/hooks/graph/__generated__/StateOfApplicabilityGraphNodeQuery.graphql";
import type { StateOfApplicabilityDetailPageExportMutation } from "./__generated__/StateOfApplicabilityDetailPageExportMutation.graphql";
import {
  StateOfApplicabilityConnectionKey,
  stateOfApplicabilityNodeQuery,
  useDeleteStateOfApplicability,
  updateStateOfApplicabilityMutation,
} from "/hooks/graph/StateOfApplicabilityGraph";
import { use, useState } from "react";
import { PermissionsContext } from "/providers/PermissionsContext";
import { usePageTitle } from "@probo/hooks";
import { formatDate, validateSnapshotConsistency } from "@probo/helpers";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { z } from "zod";
import StateOfApplicabilityControlsTab from "./tabs/StateOfApplicabilityControlsTab";
import { SnapshotBanner } from "/components/SnapshotBanner";

const exportStateOfApplicabilityPDFMutation = graphql`
  mutation StateOfApplicabilityDetailPageExportMutation($input: ExportStateOfApplicabilityPDFInput!) {
    exportStateOfApplicabilityPDF(input: $input) {
      data
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<StateOfApplicabilityGraphNodeQuery>;
};

export default function StateOfApplicabilityDetailPage(props: Props) {
  const { stateOfApplicabilityId, snapshotId } = useParams<{ stateOfApplicabilityId: string; snapshotId?: string }>();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery(stateOfApplicabilityNodeQuery, props.queryRef);
  const stateOfApplicability = data.node;
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const { isAuthorized } = use(PermissionsContext);
  const isSnapshotMode = Boolean(snapshotId);

  if (!stateOfApplicabilityId || !stateOfApplicability) {
    throw new Error(
      "Cannot load state of applicability detail page without stateOfApplicabilityId parameter",
    );
  }

  validateSnapshotConsistency(stateOfApplicability, snapshotId);

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    StateOfApplicabilityConnectionKey,
  );

  const deleteStateOfApplicability = useDeleteStateOfApplicability(
    stateOfApplicability,
    connectionId,
    () => navigate(`/organizations/${organizationId}/states-of-applicability`),
  );

  usePageTitle(stateOfApplicability.name || __("State of Applicability"));

  const [isEditingName, setIsEditingName] = useState(false);
  const [updateStateOfApplicability, isUpdating] = useMutationWithToasts(
    updateStateOfApplicabilityMutation,
    {
      successMessage: __("State of Applicability updated successfully."),
      errorMessage: __("Failed to update State of Applicability"),
    }
  );

  const canUpdate = !isSnapshotMode && isAuthorized("StateOfApplicability", "updateStateOfApplicability");
  const canDelete = !isSnapshotMode && isAuthorized("StateOfApplicability", "deleteStateOfApplicability");

  const [exportStateOfApplicabilityPDF, isExporting] = useMutationWithToasts<StateOfApplicabilityDetailPageExportMutation>(
    exportStateOfApplicabilityPDFMutation,
    {
      successMessage: __("State of Applicability exported successfully."),
      errorMessage: __("Failed to export State of Applicability"),
    }
  );

  const handleExport = () => {
    if (!stateOfApplicability.id) return;

    exportStateOfApplicabilityPDF({
      variables: {
        input: {
          stateOfApplicabilityId: stateOfApplicability.id,
        },
      },
      onCompleted: (data) => {
        if (data.exportStateOfApplicabilityPDF?.data) {
          const link = window.document.createElement("a");
          link.href = data.exportStateOfApplicabilityPDF.data;
          link.download = `${stateOfApplicability.name || "state-of-applicability"}.pdf`;
          window.document.body.appendChild(link);
          link.click();
          window.document.body.removeChild(link);
        }
      },
    });
  };

  const updateSchema = z.object({
    name: z.string().min(1, __("Name is required")),
    description: z.string().optional(),
  });

  const { register, handleSubmit, reset, watch } = useFormWithSchema(
    updateSchema,
    {
      defaultValues: {
        name: stateOfApplicability.name || "",
        description: stateOfApplicability.description || "",
      },
    }
  );

  const name = watch("name");
  const description = watch("description");
  const isDirty = name !== (stateOfApplicability.name || "") ||
                  description !== (stateOfApplicability.description || "");

  const handleUpdate = handleSubmit((data) => {
    if (!stateOfApplicability.id) return;

    updateStateOfApplicability({
      variables: {
        input: {
          id: stateOfApplicability.id,
          name: data.name,
          description: data.description || null,
        },
      },
      onSuccess: () => {
        setIsEditingName(false);
        reset({ name: data.name, description: data.description || "" });
      },
    });
  });

  const handleCancelEdit = () => {
    setIsEditingName(false);
    reset({
      name: stateOfApplicability.name || "",
      description: stateOfApplicability.description || "",
    });
  };

  const listUrl = snapshotId
    ? `/organizations/${organizationId}/snapshots/${snapshotId}/states-of-applicability`
    : `/organizations/${organizationId}/states-of-applicability`;

  return (
    <div className="space-y-6">
      {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
      <Breadcrumb
        items={[
          {
            label: __("States of Applicability"),
            to: listUrl,
          },
          {
            label: stateOfApplicability.name || __("State of Applicability detail"),
          },
        ]}
      />

      <PageHeader
        title={
          isEditingName && canUpdate ? (
            <div className="flex items-center gap-2">
              <Input
                {...register("name")}
                variant="title"
                className="flex-1"
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === "Escape") {
                    handleCancelEdit();
                  }
                  if (e.key === "Enter" && e.ctrlKey) {
                    handleUpdate();
                  }
                }}
              />
              <Button
                variant="quaternary"
                icon={IconCheckmark1}
                onClick={handleUpdate}
                disabled={isUpdating || !isDirty}
              />
              <Button
                variant="quaternary"
                icon={IconCrossLargeX}
                onClick={handleCancelEdit}
              />
            </div>
          ) : (
            <div className="flex items-center gap-2">
              <span>{stateOfApplicability.name || ""}</span>
              {canUpdate && (
                <Button
                  variant="quaternary"
                  icon={IconPencil}
                  onClick={() => setIsEditingName(true)}
                />
              )}
            </div>
          )
        }
      >
        <Button
          variant="secondary"
          icon={IconArrowDown}
          onClick={handleExport}
          disabled={isExporting}
        >
          {__("Export")}
        </Button>
        {canDelete && (
          <ActionDropdown variant="secondary">
            <DropdownItem variant="danger" icon={IconTrashCan} onClick={deleteStateOfApplicability}>
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </PageHeader>

      <div className="space-y-6">
        <div className="space-y-4">
          <h2 className="text-base font-medium">{__("Details")}</h2>
          <Card className="space-y-4" padded>
            {isEditingName && canUpdate ? (
              <Field
                label={__("Description")}
                {...register("description")}
                type="textarea"
              />
            ) : (
              <>
                <div>
                  <div className="text-xs text-txt-tertiary font-semibold mb-1">
                    {__("Description")}
                  </div>
                  <div className="text-sm text-txt-primary whitespace-pre-wrap">
                    {stateOfApplicability.description || "-"}
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <div className="text-xs text-txt-tertiary font-semibold mb-1">
                      {__("Created at")}
                    </div>
                    <div className="text-sm text-txt-primary">
                      {formatDate(stateOfApplicability.createdAt)}
                    </div>
                  </div>
                  <div>
                    <div className="text-xs text-txt-tertiary font-semibold mb-1">
                      {__("Updated at")}
                    </div>
                    <div className="text-sm text-txt-primary">
                      {formatDate(stateOfApplicability.updatedAt)}
                    </div>
                  </div>
                </div>
              </>
            )}
          </Card>
          {isEditingName && canUpdate && isDirty && (
            <div className="flex justify-end gap-2">
              <Button
                variant="secondary"
                onClick={handleCancelEdit}
              >
                {__("Cancel")}
              </Button>
              <Button
                onClick={handleUpdate}
                disabled={isUpdating}
              >
                {__("Save")}
              </Button>
            </div>
          )}
        </div>

        {stateOfApplicability.id && (
          <div className="space-y-4">
            <h2 className="text-base font-medium">{__("Controls")}</h2>
            <StateOfApplicabilityControlsTab
              stateOfApplicability={stateOfApplicability as typeof stateOfApplicability & { id: string }}
              isSnapshotMode={isSnapshotMode}
            />
          </div>
        )}
      </div>
    </div>
  );
}
