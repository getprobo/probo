import { formatDate, validateSnapshotConsistency } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Breadcrumb,
  Button,
  Card,
  DropdownItem,
  IconArrowDown,
  IconCheckmark1,
  IconCrossLargeX,
  IconPencil,
  IconTrashCan,
  Input,
  PageHeader,
} from "@probo/ui";
import { Suspense, useState } from "react";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate, useParams } from "react-router";
import { z } from "zod";

import type { StateOfApplicabilityDetailPageExportMutation } from "/__generated__/core/StateOfApplicabilityDetailPageExportMutation.graphql";
import type { StateOfApplicabilityGraphNodeQuery } from "/__generated__/core/StateOfApplicabilityGraphNodeQuery.graphql";
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import { SnapshotBanner } from "/components/SnapshotBanner";
import {
  StateOfApplicabilityConnectionKey,
  stateOfApplicabilityNodeQuery,
  updateStateOfApplicabilityMutation,
  useDeleteStateOfApplicability,
} from "/hooks/graph/StateOfApplicabilityGraph";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useOrganizationId } from "/hooks/useOrganizationId";

import StateOfApplicabilityControlsTab from "./tabs/StateOfApplicabilityControlsTab";

const exportStateOfApplicabilityPDFMutation = graphql`
    mutation StateOfApplicabilityDetailPageExportMutation(
        $input: ExportStateOfApplicabilityPDFInput!
    ) {
        exportStateOfApplicabilityPDF(input: $input) {
            data
        }
    }
`;

type Props = {
  queryRef: PreloadedQuery<StateOfApplicabilityGraphNodeQuery>;
};

export default function StateOfApplicabilityDetailPage(props: Props) {
  const { stateOfApplicabilityId, snapshotId } = useParams<{
    stateOfApplicabilityId: string;
    snapshotId?: string;
  }>();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery(
    stateOfApplicabilityNodeQuery,
    props.queryRef,
  );
  const stateOfApplicability = data.node;
  const { __ } = useTranslate();
  const navigate = useNavigate();
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
    () =>
      void navigate(
        `/organizations/${organizationId}/states-of-applicability`,
      ),
  );

  usePageTitle(stateOfApplicability.name || __("State of Applicability"));

  const [isEditingName, setIsEditingName] = useState(false);
  const [isEditingOwner, setIsEditingOwner] = useState(false);
  const [updateStateOfApplicability, isUpdating] = useMutationWithToasts(
    updateStateOfApplicabilityMutation,
    {
      successMessage: __("State of Applicability updated successfully."),
      errorMessage: __("Failed to update State of Applicability"),
    },
  );

  const canUpdate = !isSnapshotMode && stateOfApplicability.canUpdate;
  const canDelete = !isSnapshotMode && stateOfApplicability.canDelete;

  const [exportStateOfApplicabilityPDF, isExporting]
    = useMutationWithToasts<StateOfApplicabilityDetailPageExportMutation>(
      exportStateOfApplicabilityPDFMutation,
      {
        successMessage: __(
          "State of Applicability exported successfully.",
        ),
        errorMessage: __("Failed to export State of Applicability"),
      },
    );

  const handleExport = async () => {
    if (!stateOfApplicability.id) return;

    await exportStateOfApplicabilityPDF({
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

  const nameSchema = z.object({
    name: z.string().min(1, __("Name is required")),
  });

  const ownerSchema = z.object({
    ownerId: z.string().min(1, __("Owner is required")),
  });

  const {
    register: registerName,
    handleSubmit: handleSubmitName,
    reset: resetName,
  } = useFormWithSchema(nameSchema, {
    defaultValues: {
      name: stateOfApplicability.name || "",
    },
  });

  const {
    control: controlOwner,
    handleSubmit: handleSubmitOwner,
    reset: resetOwner,
  } = useFormWithSchema(ownerSchema, {
    defaultValues: {
      ownerId: stateOfApplicability.owner?.id || "",
    },
  });

  const handleUpdateName = handleSubmitName(async (data) => {
    if (!stateOfApplicability.id) return;

    await updateStateOfApplicability({
      variables: {
        input: {
          id: stateOfApplicability.id,
          name: data.name,
        },
      },
      onSuccess: () => {
        setIsEditingName(false);
        resetName({ name: data.name });
      },
    });
  });

  const handleUpdateOwner = handleSubmitOwner(async (data) => {
    if (!stateOfApplicability.id) return;

    await updateStateOfApplicability({
      variables: {
        input: {
          id: stateOfApplicability.id,
          ownerId: data.ownerId,
        },
      },
      onSuccess: () => {
        setIsEditingOwner(false);
        resetOwner({ ownerId: data.ownerId });
      },
    });
  });

  const handleCancelNameEdit = () => {
    setIsEditingName(false);
    resetName({
      name: stateOfApplicability.name || "",
    });
  };

  const handleCancelOwnerEdit = () => {
    setIsEditingOwner(false);
    resetOwner({
      ownerId: stateOfApplicability.owner?.id || "",
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
            label:
                            stateOfApplicability.name
                            || __("State of Applicability detail"),
          },
        ]}
      />

      <PageHeader
        title={
          isEditingName && canUpdate
            ? (
                <div className="flex items-center gap-2">
                  <Input
                    {...registerName("name")}
                    variant="title"
                    className="flex-1"
                    autoFocus
                    onKeyDown={(e) => {
                      if (e.key === "Escape") {
                        handleCancelNameEdit();
                      }
                      if (e.key === "Enter" && e.ctrlKey) {
                        void handleUpdateName();
                      }
                    }}
                  />
                  <Button
                    variant="quaternary"
                    icon={IconCheckmark1}
                    onClick={() => void handleUpdateName()}
                    disabled={isUpdating}
                  />
                  <Button
                    variant="quaternary"
                    icon={IconCrossLargeX}
                    onClick={handleCancelNameEdit}
                  />
                </div>
              )
            : (
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
        {stateOfApplicability.canExport && (
          <Button
            variant="secondary"
            icon={IconArrowDown}
            onClick={() => void handleExport()}
            disabled={isExporting}
          >
            {__("Export")}
          </Button>
        )}
        {canDelete && (
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={() => void deleteStateOfApplicability()}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </PageHeader>

      <div className="space-y-6">
        <div className="space-y-4">
          <h2 className="text-base font-medium">{__("Details")}</h2>
          <Card className="space-y-4" padded>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="text-xs text-txt-tertiary font-semibold mb-1">
                  {__("Owner")}
                </div>
                {isEditingOwner && canUpdate
                  ? (
                      <div className="flex items-center gap-2">
                        <Suspense
                          fallback={
                            <div>{__("Loading...")}</div>
                          }
                        >
                          <PeopleSelectField
                            organizationId={organizationId}
                            control={controlOwner}
                            name="ownerId"
                          />
                        </Suspense>
                        <Button
                          variant="quaternary"
                          icon={IconCheckmark1}
                          onClick={() => void handleUpdateOwner()}
                          disabled={isUpdating}
                        />
                        <Button
                          variant="quaternary"
                          icon={IconCrossLargeX}
                          onClick={handleCancelOwnerEdit}
                        />
                      </div>
                    )
                  : (
                      <div className="flex items-center gap-2">
                        <div className="text-sm text-txt-primary">
                          {stateOfApplicability.owner
                            ?.fullName || "-"}
                        </div>
                        {canUpdate && (
                          <Button
                            variant="quaternary"
                            icon={IconPencil}
                            onClick={() =>
                              setIsEditingOwner(true)}
                          />
                        )}
                      </div>
                    )}
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
          </Card>
        </div>

        {stateOfApplicability.id && (
          <div className="space-y-4">
            <h2 className="text-base font-medium">
              {__("Controls")}
            </h2>
            <StateOfApplicabilityControlsTab
              stateOfApplicability={
                stateOfApplicability as typeof stateOfApplicability & {
                  id: string;
                }
              }
              isSnapshotMode={isSnapshotMode}
            />
          </div>
        )}
      </div>
    </div>
  );
}
