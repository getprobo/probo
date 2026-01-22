import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Button,
  DropdownItem,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { Suspense, useMemo, useRef } from "react";
import { graphql, useRefetchableFragment } from "react-relay";

import type { StateOfApplicabilityControlsTabFragment$key } from "#/__generated__/core/StateOfApplicabilityControlsTabFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  EditControlDialog,
  type EditControlDialogRef,
} from "../dialogs/EditControlDialog";
import {
  LinkControlDialog,
  type LinkControlDialogRef,
} from "../dialogs/LinkControlDialog";

export const controlsFragment = graphql`
    fragment StateOfApplicabilityControlsTabFragment on StateOfApplicability
    @refetchable(queryName: "StateOfApplicabilityControlsTabRefetchQuery") {
        id
        controlsInfo: controls(first: 0) {
            totalCount
        }

        canCreateStateOfApplicabilityControlMapping: permission(
            action: "core:state-of-applicability-control-mapping:create"
        )
        canDeleteStateOfApplicabilityControlMapping: permission(
            action: "core:state-of-applicability-control-mapping:delete"
        )

        availableControls {
            controlId
            sectionTitle
            name
            frameworkId
            frameworkName
            organizationId
            stateOfApplicabilityId
            applicability
            justification
            bestPractice
            regulatory
            contractual
            riskAssessment
        }
    }
`;

const unlinkControlMutation = graphql`
    mutation StateOfApplicabilityControlsTabUnlinkMutation(
        $input: DeleteStateOfApplicabilityControlMappingInput!
    ) {
        deleteStateOfApplicabilityControlMapping(input: $input) {
            deletedStateOfApplicabilityId
            deletedControlId
            deletedStateOfApplicabilityControlId
        }
    }
`;

export default function StateOfApplicabilityControlsTab({
  stateOfApplicability,
  isSnapshotMode = false,
}: {
  stateOfApplicability: StateOfApplicabilityControlsTabFragment$key & {
    id: string;
  };
  isSnapshotMode?: boolean;
}) {
  const { __ } = useTranslate();
  const [data, refetch] = useRefetchableFragment(
    controlsFragment,
    stateOfApplicability,
  );
  const organizationId = useOrganizationId();
  const manageDialogRef = useRef<LinkControlDialogRef>(null);
  const editDialogRef = useRef<EditControlDialogRef>(null);

  const linkedControls = useMemo(
    () =>
      (data.availableControls || []).filter(
        c => c.stateOfApplicabilityId !== null,
      ),
    [data.availableControls],
  );

  const [unlinkControl, isUnlinking] = useMutationWithToasts(
    unlinkControlMutation,
    {
      successMessage: __("Control removed successfully."),
      errorMessage: __("Failed to remove control"),
    },
  );

  const canLink
    = !isSnapshotMode && data.canCreateStateOfApplicabilityControlMapping;
  const canUnlink
    = !isSnapshotMode && data.canDeleteStateOfApplicabilityControlMapping;

  const handleOpenManageDialog = () => {
    manageDialogRef.current?.open(data.id, () => {
      refetch({}, { fetchPolicy: "store-and-network" });
    });
  };

  const handleOpenEditDialog = (control: {
    controlId: string;
    sectionTitle: string;
    name: string;
    frameworkName: string;
    applicability: boolean;
    justification: string | null;
  }) => {
    editDialogRef.current?.open({
      stateOfApplicabilityId: data.id,
      controlId: control.controlId,
      sectionTitle: control.sectionTitle,
      name: control.name,
      frameworkName: control.frameworkName,
      applicability: control.applicability,
      justification: control.justification,
    });
  };

  const handleUnlink = async (controlId: string) => {
    await unlinkControl({
      variables: {
        input: {
          stateOfApplicabilityId: data.id,
          controlId,
        },
      },
      onSuccess: () => {
        refetch({}, { fetchPolicy: "network-only" });
      },
    });
  };

  return (
    <>
      <div className="space-y-4">
        {canLink && (
          <div className="flex justify-end">
            <Button
              icon={IconPlusLarge}
              onClick={handleOpenManageDialog}
            >
              {__("Add Controls")}
            </Button>
          </div>
        )}

        <Table>
          <Thead>
            <Tr>
              <Th className="w-32">{__("Framework")}</Th>
              <Th>{__("Control")}</Th>
              <Th className="w-28 text-center">
                {__("Applicability")}
              </Th>
              <Th className="min-w-48">{__("Justification")}</Th>
              <Th className="w-24 text-center">
                {__("Regulatory")}
              </Th>
              <Th className="w-24 text-center">
                {__("Contractual")}
              </Th>
              <Th className="w-32 text-center">
                {__("Best Practice")}
              </Th>
              <Th className="w-36 text-center">
                {__("Risk Assessment")}
              </Th>
              {(canLink || canUnlink) && (
                <Th className="w-12"></Th>
              )}
            </Tr>
          </Thead>
          <Tbody>
            {linkedControls.length === 0 && (
              <Tr>
                <Td
                  colSpan={canLink || canUnlink ? 9 : 8}
                  className="text-center text-txt-secondary py-12"
                >
                  {__("No controls linked")}
                </Td>
              </Tr>
            )}
            {linkedControls.map(control => (
              <Tr
                key={control.controlId}
                to={`/organizations/${organizationId}/frameworks/${control.frameworkId}/controls/${control.controlId}`}
              >
                <Td className="font-medium text-txt-secondary">
                  {control.frameworkName}
                </Td>
                <Td>
                  <div className="space-y-1">
                    <div className="text-xs font-medium text-txt-tertiary">
                      {control.sectionTitle}
                    </div>
                    <div className="text-sm">
                      {control.name}
                    </div>
                  </div>
                </Td>
                <Td>
                  <div className="flex justify-center">
                    {control.applicability !== null
                      ? (
                          <Badge
                            variant={
                              control.applicability
                                ? "success"
                                : "danger"
                            }
                            size="sm"
                          >
                            {control.applicability
                              ? __("Yes")
                              : __("No")}
                          </Badge>
                        )
                      : (
                          <span className="text-txt-tertiary">
                            -
                          </span>
                        )}
                  </div>
                </Td>
                <Td>
                  <div className="text-sm text-txt-secondary line-clamp-2">
                    {control.justification || (
                      <span className="text-txt-tertiary italic">
                        -
                      </span>
                    )}
                  </div>
                </Td>
                <Td>
                  <div className="flex justify-center">
                    <Badge
                      variant={
                        control.regulatory
                          ? "success"
                          : "danger"
                      }
                      size="sm"
                    >
                      {control.regulatory
                        ? __("Yes")
                        : __("No")}
                    </Badge>
                  </div>
                </Td>
                <Td>
                  <div className="flex justify-center">
                    <Badge
                      variant={
                        control.contractual
                          ? "success"
                          : "danger"
                      }
                      size="sm"
                    >
                      {control.contractual
                        ? __("Yes")
                        : __("No")}
                    </Badge>
                  </div>
                </Td>
                <Td>
                  <div className="flex justify-center">
                    <Badge
                      variant={
                        control.bestPractice
                          ? "success"
                          : "danger"
                      }
                      size="sm"
                    >
                      {control.bestPractice
                        ? __("Yes")
                        : __("No")}
                    </Badge>
                  </div>
                </Td>
                <Td>
                  <div className="flex justify-center">
                    <Badge
                      variant={
                        control.riskAssessment
                          ? "success"
                          : "danger"
                      }
                      size="sm"
                    >
                      {control.riskAssessment
                        ? __("Yes")
                        : __("No")}
                    </Badge>
                  </div>
                </Td>
                {(canLink || canUnlink) && (
                  <Td noLink className="text-end">
                    <ActionDropdown>
                      {canLink && (
                        <DropdownItem
                          icon={IconPencil}
                          onClick={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            if (
                              typeof control.applicability
                              === "boolean"
                            ) {
                              handleOpenEditDialog(
                                {
                                  controlId: control.controlId,
                                  sectionTitle: control.sectionTitle,
                                  name: control.name,
                                  frameworkName: control.frameworkName,
                                  applicability: control.applicability,
                                  justification: control.justification ?? null,
                                },
                              );
                            }
                          }}
                        >
                          {__("Edit")}
                        </DropdownItem>
                      )}
                      {canUnlink && (
                        <DropdownItem
                          icon={IconTrashCan}
                          variant="danger"
                          onClick={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            void handleUnlink(control.controlId);
                          }}
                          disabled={isUnlinking}
                        >
                          {__("Remove")}
                        </DropdownItem>
                      )}
                    </ActionDropdown>
                  </Td>
                )}
              </Tr>
            ))}
          </Tbody>
        </Table>
      </div>

      <Suspense fallback={null}>
        <LinkControlDialog ref={manageDialogRef} />
        <EditControlDialog
          ref={editDialogRef}
          onSuccess={() => {
            refetch({}, { fetchPolicy: "network-only" });
          }}
        />
      </Suspense>
    </>
  );
}
