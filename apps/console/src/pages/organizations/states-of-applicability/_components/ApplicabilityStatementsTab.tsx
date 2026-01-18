import { graphql, useRefetchableFragment } from "react-relay";
import {
    Badge,
    Table,
    Tbody,
    Td,
    Th,
    Thead,
    Tr,
    Button,
    IconPlusLarge,
    IconPencil,
    IconTrashCan,
    ActionDropdown,
    DropdownItem,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { Suspense, useMemo, useRef } from "react";
import {
    ManageApplicabilityStatementsDialog,
    type ManageApplicabilityStatementsDialogRef,
} from "./ManageApplicabilityStatementsDialog";
import {
    EditApplicabilityStatementDialog,
    type EditApplicabilityStatementDialogRef,
} from "./EditApplicabilityStatementDialog";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type { ApplicabilityStatementsTabFragment$key } from "/__generated__/core/ApplicabilityStatementsTabFragment.graphql";

export const applicabilityStatementsTabFragment = graphql`
    fragment ApplicabilityStatementsTabFragment on StateOfApplicability
    @refetchable(queryName: "ApplicabilityStatementsTabRefetchQuery") {
        id
        applicabilityStatementsInfo: applicabilityStatements(first: 0) {
            totalCount
        }

        canCreateApplicabilityStatement: permission(
            action: "core:state-of-applicability-control-mapping:create"
        )
        canDeleteApplicabilityStatement: permission(
            action: "core:state-of-applicability-control-mapping:delete"
        )

        availableControls {
            controlId
            sectionTitle
            name
            frameworkId
            frameworkName
            organizationId
            applicabilityStatementId
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

const unlinkApplicabilityStatementMutation = graphql`
    mutation ApplicabilityStatementsTabUnlinkMutation(
        $input: DeleteApplicabilityStatementInput!
    ) {
        deleteApplicabilityStatement(input: $input) {
            deletedApplicabilityStatementId
        }
    }
`;

export default function ApplicabilityStatementsTab({
    fKey,
    isSnapshotMode = false,
}: {
    fKey: ApplicabilityStatementsTabFragment$key;
    isSnapshotMode?: boolean;
}) {
    const { __ } = useTranslate();
    const [data, refetch] = useRefetchableFragment(
        applicabilityStatementsTabFragment,
        fKey,
    );
    const organizationId = useOrganizationId();
    const manageDialogRef =
        useRef<ManageApplicabilityStatementsDialogRef>(null);
    const editDialogRef = useRef<EditApplicabilityStatementDialogRef>(null);

    const linkedStatements = useMemo(
        () =>
            (data.availableControls || []).filter(
                (c) => c.stateOfApplicabilityId !== null,
            ),
        [data.availableControls],
    );

    const [unlinkStatement, isUnlinking] = useMutationWithToasts(
        unlinkApplicabilityStatementMutation,
        {
            successMessage: __("Statement removed successfully."),
            errorMessage: __("Failed to remove statement"),
        },
    );

    const canLink = !isSnapshotMode && data.canCreateApplicabilityStatement;
    const canUnlink = !isSnapshotMode && data.canDeleteApplicabilityStatement;

    const handleOpenManageDialog = () => {
        manageDialogRef.current?.open(data.id, () => {
            refetch({}, { fetchPolicy: "store-and-network" });
        });
    };

    const handleOpenEditDialog = (statement: {
        applicabilityStatementId: string;
        sectionTitle: string;
        name: string;
        frameworkName: string;
        applicability: boolean;
        justification: string | null;
    }) => {
        editDialogRef.current?.open({
            applicabilityStatementId: statement.applicabilityStatementId,
            stateOfApplicabilityId: data.id,
            sectionTitle: statement.sectionTitle,
            name: statement.name,
            frameworkName: statement.frameworkName,
            applicability: statement.applicability,
            justification: statement.justification,
        });
    };

    const handleUnlink = (applicabilityStatementId: string) => {
        unlinkStatement({
            variables: {
                input: {
                    applicabilityStatementId,
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
                            {__("Add Statements")}
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
                        {linkedStatements.length === 0 && (
                            <Tr>
                                <Td
                                    colSpan={canLink || canUnlink ? 9 : 8}
                                    className="text-center text-txt-secondary py-12"
                                >
                                    {__("No statements linked")}
                                </Td>
                            </Tr>
                        )}
                        {linkedStatements.map((statement) => (
                            <ApplicabilityStatementRow
                                key={statement.controlId}
                                statement={statement}
                                organizationId={organizationId}
                                canLink={canLink}
                                canUnlink={canUnlink}
                                isUnlinking={isUnlinking}
                                onEdit={handleOpenEditDialog}
                                onUnlink={handleUnlink}
                            />
                        ))}
                    </Tbody>
                </Table>
            </div>

            <Suspense fallback={null}>
                <ManageApplicabilityStatementsDialog ref={manageDialogRef} />
                <EditApplicabilityStatementDialog
                    ref={editDialogRef}
                    onSuccess={() => {
                        refetch({}, { fetchPolicy: "network-only" });
                    }}
                />
            </Suspense>
        </>
    );
}

function ApplicabilityStatementRow({
    statement,
    organizationId,
    canLink,
    canUnlink,
    isUnlinking,
    onEdit,
    onUnlink,
}: {
    statement: {
        controlId: string;
        sectionTitle: string;
        name: string;
        frameworkId: string;
        frameworkName: string;
        applicabilityStatementId: string | null | undefined;
        applicability: boolean | null | undefined;
        justification: string | null | undefined;
        bestPractice: boolean;
        regulatory: boolean;
        contractual: boolean;
        riskAssessment: boolean;
    };
    organizationId: string;
    canLink: boolean;
    canUnlink: boolean;
    isUnlinking: boolean;
    onEdit: (statement: {
        applicabilityStatementId: string;
        sectionTitle: string;
        name: string;
        frameworkName: string;
        applicability: boolean;
        justification: string | null;
    }) => void;
    onUnlink: (applicabilityStatementId: string) => void;
}) {
    const { __ } = useTranslate();

    return (
        <Tr
            to={`/organizations/${organizationId}/frameworks/${statement.frameworkId}/controls/${statement.controlId}`}
        >
            <Td className="font-medium text-txt-secondary">
                {statement.frameworkName}
            </Td>
            <Td>
                <div className="space-y-1">
                    <div className="text-xs font-medium text-txt-tertiary">
                        {statement.sectionTitle}
                    </div>
                    <div className="text-sm">{statement.name}</div>
                </div>
            </Td>
            <Td>
                <div className="flex justify-center">
                    {statement.applicability !== null ? (
                        <Badge
                            variant={
                                statement.applicability ? "success" : "danger"
                            }
                            size="sm"
                        >
                            {statement.applicability ? __("Yes") : __("No")}
                        </Badge>
                    ) : (
                        <span className="text-txt-tertiary">-</span>
                    )}
                </div>
            </Td>
            <Td>
                <div className="text-sm text-txt-secondary line-clamp-2">
                    {statement.justification || (
                        <span className="text-txt-tertiary italic">-</span>
                    )}
                </div>
            </Td>
            <Td>
                <div className="flex justify-center">
                    <Badge
                        variant={statement.regulatory ? "success" : "danger"}
                        size="sm"
                    >
                        {statement.regulatory ? __("Yes") : __("No")}
                    </Badge>
                </div>
            </Td>
            <Td>
                <div className="flex justify-center">
                    <Badge
                        variant={statement.contractual ? "success" : "danger"}
                        size="sm"
                    >
                        {statement.contractual ? __("Yes") : __("No")}
                    </Badge>
                </div>
            </Td>
            <Td>
                <div className="flex justify-center">
                    <Badge
                        variant={statement.bestPractice ? "success" : "danger"}
                        size="sm"
                    >
                        {statement.bestPractice ? __("Yes") : __("No")}
                    </Badge>
                </div>
            </Td>
            <Td>
                <div className="flex justify-center">
                    <Badge
                        variant={
                            statement.riskAssessment ? "success" : "danger"
                        }
                        size="sm"
                    >
                        {statement.riskAssessment ? __("Yes") : __("No")}
                    </Badge>
                </div>
            </Td>
            {(canLink || canUnlink) && (
                <Td noLink className="text-end">
                    <ActionDropdown>
                        {canLink && statement.applicabilityStatementId && (
                            <DropdownItem
                                icon={IconPencil}
                                onClick={(e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    if (
                                        typeof statement.applicability ===
                                        "boolean"
                                    ) {
                                        onEdit({
                                            applicabilityStatementId:
                                                statement.applicabilityStatementId!,
                                            sectionTitle:
                                                statement.sectionTitle,
                                            name: statement.name,
                                            frameworkName:
                                                statement.frameworkName,
                                            applicability:
                                                statement.applicability,
                                            justification:
                                                statement.justification ?? null,
                                        });
                                    }
                                }}
                            >
                                {__("Edit")}
                            </DropdownItem>
                        )}
                        {canUnlink && statement.applicabilityStatementId && (
                            <DropdownItem
                                icon={IconTrashCan}
                                variant="danger"
                                onClick={(e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    onUnlink(
                                        statement.applicabilityStatementId!,
                                    );
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
    );
}
