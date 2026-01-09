import { graphql, useRefetchableFragment } from "react-relay";
import { Badge, Table, Tbody, Td, Th, Thead, Tr, Button, IconPlusLarge, IconPencil, IconTrashCan, ActionDropdown, DropdownItem } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { StateOfApplicabilityControlsTabFragment$key } from "./__generated__/StateOfApplicabilityControlsTabFragment.graphql";
import { Suspense, useMemo, useRef } from "react";
import { LinkControlDialog, type LinkControlDialogRef } from "../dialogs/LinkControlDialog";
import { EditControlDialog, type EditControlDialogRef } from "../dialogs/EditControlDialog";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { use } from "react";
import { PermissionsContext } from "/providers/PermissionsContext";

export const controlsFragment = graphql`
  fragment StateOfApplicabilityControlsTabFragment on StateOfApplicability
  @refetchable(queryName: "StateOfApplicabilityControlsTabRefetchQuery") {
    id
    availableControls {
      controlId
      sectionTitle
      name
      frameworkId
      frameworkName
      organizationId
      stateOfApplicabilityId
      state
      exclusionJustification
    }
  }
`;

const unlinkControlMutation = graphql`
  mutation StateOfApplicabilityControlsTabUnlinkMutation($input: UnlinkStateOfApplicabilityControlInput!) {
    unlinkStateOfApplicabilityControl(input: $input) {
      deletedControlId
    }
  }
`;

const stateLabels: Record<string, string> = {
  IMPLEMENTED: "Implemented",
  NOT_IMPLEMENTED: "Not Implemented",
  EXCLUDED: "Excluded",
};

export default function StateOfApplicabilityControlsTab({
  stateOfApplicability,
  isSnapshotMode = false,
}: {
  stateOfApplicability: StateOfApplicabilityControlsTabFragment$key & { id: string };
  isSnapshotMode?: boolean;
}) {
  const { __ } = useTranslate();
  const [data, refetch] = useRefetchableFragment(controlsFragment, stateOfApplicability);
  const organizationId = useOrganizationId();
  const manageDialogRef = useRef<LinkControlDialogRef>(null);
  const editDialogRef = useRef<EditControlDialogRef>(null);
  const { isAuthorized } = use(PermissionsContext);

  const linkedControls = useMemo(
    () => (data.availableControls || []).filter((c) => c.stateOfApplicabilityId !== null),
    [data.availableControls]
  );

  const [unlinkControl, isUnlinking] = useMutationWithToasts(
    unlinkControlMutation,
    {
      successMessage: __("Control removed successfully."),
      errorMessage: __("Failed to remove control"),
    },
  );

  const canLink = !isSnapshotMode && isAuthorized("StateOfApplicability", "updateStateOfApplicability");
  const canUnlink = !isSnapshotMode && isAuthorized("StateOfApplicability", "updateStateOfApplicability");

  const handleOpenManageDialog = () => {
    manageDialogRef.current?.open(data.id, () => {
      refetch({}, { fetchPolicy: "network-only" });
    });
  };

  const handleOpenEditDialog = (control: {
    controlId: string;
    sectionTitle: string;
    name: string;
    frameworkName: string;
    state: string | null;
    exclusionJustification: string | null;
  }) => {
    editDialogRef.current?.open({
      stateOfApplicabilityId: data.id,
      controlId: control.controlId,
      sectionTitle: control.sectionTitle,
      name: control.name,
      frameworkName: control.frameworkName,
      state: control.state,
      exclusionJustification: control.exclusionJustification,
    });
  };

  const handleUnlink = (controlId: string) => {
    unlinkControl({
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
              {__("Manage Controls")}
            </Button>
          </div>
        )}

        <Table>
          <Thead>
            <Tr>
              <Th>{__("Framework")}</Th>
              <Th>{__("Reference")}</Th>
              <Th>{__("Name")}</Th>
              <Th>{__("State")}</Th>
              <Th>{__("Exclusion Justification")}</Th>
              {(canLink || canUnlink) && <Th></Th>}
            </Tr>
          </Thead>
          <Tbody>
            {linkedControls.length === 0 && (
              <Tr>
                <Td colSpan={canLink || canUnlink ? 6 : 5} className="text-center text-txt-secondary">
                  {__("No controls linked")}
                </Td>
              </Tr>
            )}
            {linkedControls.map((control) => (
              <Tr
                key={control.controlId}
                to={`/organizations/${organizationId}/frameworks/${control.frameworkId}/controls/${control.controlId}`}
              >
                <Td>{control.frameworkName}</Td>
                <Td>
                  <Badge size="md">{control.sectionTitle}</Badge>
                </Td>
                <Td>{control.name}</Td>
                <Td>
                  {control.state ? __(stateLabels[control.state] || control.state) : "-"}
                </Td>
                <Td>
                  {control.exclusionJustification || "-"}
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
                            handleOpenEditDialog({
                              controlId: control.controlId,
                              sectionTitle: control.sectionTitle,
                              name: control.name,
                              frameworkName: control.frameworkName,
                              state: control.state ?? null,
                              exclusionJustification: control.exclusionJustification ?? null,
                            });
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
                            handleUnlink(control.controlId);
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
