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
import { Suspense, useCallback, useRef } from "react";
import { graphql, useFragment, useRelayEnvironment } from "react-relay";
import { fetchQuery } from "relay-runtime";

import type { StateOfApplicabilityControlsTabFragment$key } from "#/__generated__/core/StateOfApplicabilityControlsTabFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  AddApplicabilityStatementDialog,
  type AddApplicabilityStatementDialogRef,
} from "../dialogs/AddApplicabilityStatementDialog";
import {
  EditControlDialog,
  type EditControlDialogRef,
} from "../dialogs/EditControlDialog";

const refetchStatementsQuery = graphql`
    query StateOfApplicabilityControlsTabRefetchQuery($stateOfApplicabilityId: ID!) {
        node(id: $stateOfApplicabilityId) {
            ... on StateOfApplicability {
                ...StateOfApplicabilityControlsTabFragment
            }
        }
    }
`;

export const controlsFragment = graphql`
    fragment StateOfApplicabilityControlsTabFragment on StateOfApplicability {
        id
        organization {
            id
        }

        canCreateApplicabilityStatement: permission(
            action: "core:applicability-statement:create"
        )
        canUpdateApplicabilityStatement: permission(
            action: "core:applicability-statement:update"
        )
        canDeleteApplicabilityStatement: permission(
            action: "core:applicability-statement:delete"
        )

        applicabilityStatements(first: 1000, orderBy: { direction: ASC, field: CONTROL_SECTION_TITLE })
            @connection(key: "StateOfApplicabilityControlsTab_applicabilityStatements") {
            __id
            edges {
                node {
                    id
                    applicability
                    justification
                    control {
                        id
                        sectionTitle
                        name
                        bestPractice
                        regulatory
                        contractual
                        riskAssessment
                        framework {
                            id
                            name
                        }
                        organization {
                            id
                        }
                    }
                }
            }
        }
    }
`;

const deleteApplicabilityStatementMutation = graphql`
    mutation StateOfApplicabilityControlsTabDeleteMutation(
        $input: DeleteApplicabilityStatementInput!
        $connections: [ID!]!
    ) {
        deleteApplicabilityStatement(input: $input) {
            deletedApplicabilityStatementId @deleteEdge(connections: $connections)
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
  const data = useFragment(controlsFragment, stateOfApplicability);
  const environment = useRelayEnvironment();
  const organizationId = useOrganizationId();
  const addStatementDialogRef = useRef<AddApplicabilityStatementDialogRef>(null);
  const editDialogRef = useRef<EditControlDialogRef>(null);

  const handleDialogClose = useCallback(() => {
    if (!data.id) return;
    fetchQuery(environment, refetchStatementsQuery, {
      stateOfApplicabilityId: data.id,
    }).subscribe({});
  }, [environment, data.id]);

  const connectionId = data.applicabilityStatements?.__id;

  const linkedControls = (data.applicabilityStatements?.edges || []).map(edge => ({
    applicabilityStatementId: edge.node.id,
    controlId: edge.node.control.id,
    sectionTitle: edge.node.control.sectionTitle,
    name: edge.node.control.name,
    frameworkId: edge.node.control.framework.id,
    frameworkName: edge.node.control.framework.name,
    organizationId: edge.node.control.organization?.id,
    applicability: edge.node.applicability,
    justification: edge.node.justification,
    bestPractice: edge.node.control.bestPractice,
    regulatory: edge.node.control.regulatory,
    contractual: edge.node.control.contractual,
    riskAssessment: edge.node.control.riskAssessment,
  }));

  const [deleteApplicabilityStatement, isDeleting] = useMutationWithToasts(
    deleteApplicabilityStatementMutation,
    {
      successMessage: __("Statement removed successfully."),
      errorMessage: __("Failed to remove statement"),
    },
  );

  const canCreate
    = !isSnapshotMode && data.canCreateApplicabilityStatement;
  const canUpdate
    = !isSnapshotMode && data.canUpdateApplicabilityStatement;
  const canDelete
    = !isSnapshotMode && data.canDeleteApplicabilityStatement;

  const handleOpenAddStatementDialog = () => {
    if (!data.organization || !connectionId) return;
    addStatementDialogRef.current?.open(data.id, data.organization.id, connectionId);
  };

  const handleOpenEditDialog = (control: {
    applicabilityStatementId: string;
    sectionTitle: string;
    name: string;
    frameworkName: string;
    applicability: boolean;
    justification: string | null;
  }) => {
    editDialogRef.current?.open({
      applicabilityStatementId: control.applicabilityStatementId,
      sectionTitle: control.sectionTitle,
      name: control.name,
      frameworkName: control.frameworkName,
      applicability: control.applicability,
      justification: control.justification,
    });
  };

  const handleDelete = async (applicabilityStatementId: string) => {
    if (!connectionId) return;
    await deleteApplicabilityStatement({
      variables: {
        input: {
          applicabilityStatementId,
        },
        connections: [connectionId],
      },
    });
  };

  return (
    <>
      <div className="space-y-4">
        {canCreate && (
          <div className="flex justify-end">
            <Button
              icon={IconPlusLarge}
              onClick={handleOpenAddStatementDialog}
            >
              {__("Create Statement")}
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
              {(canUpdate || canDelete) && (
                <Th className="w-12"></Th>
              )}
            </Tr>
          </Thead>
          <Tbody>
            {linkedControls.length === 0 && (
              <Tr>
                <Td
                  colSpan={canUpdate || canDelete ? 9 : 8}
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
                {(canUpdate || canDelete) && (
                  <Td noLink className="text-end">
                    <ActionDropdown>
                      {canUpdate && control.applicabilityStatementId && (
                        <DropdownItem
                          icon={IconPencil}
                          onClick={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            if (
                              typeof control.applicability
                              === "boolean"
                              && control.applicabilityStatementId
                            ) {
                              handleOpenEditDialog(
                                {
                                  applicabilityStatementId:
                                                                        control.applicabilityStatementId,
                                  sectionTitle:
                                                                        control.sectionTitle,
                                  name: control.name,
                                  frameworkName:
                                                                        control.frameworkName,
                                  applicability:
                                                                        control.applicability,
                                  justification:
                                                                        control.justification
                                                                        ?? null,
                                },
                              );
                            }
                          }}
                        >
                          {__("Edit")}
                        </DropdownItem>
                      )}
                      {canDelete && control.applicabilityStatementId && (
                        <DropdownItem
                          icon={IconTrashCan}
                          variant="danger"
                          onClick={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            if (control.applicabilityStatementId) {
                              void handleDelete(
                                control.applicabilityStatementId,
                              );
                            }
                          }}
                          disabled={isDeleting}
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
        <AddApplicabilityStatementDialog ref={addStatementDialogRef} onClose={handleDialogClose} />
        <EditControlDialog ref={editDialogRef} />
      </Suspense>
    </>
  );
}
