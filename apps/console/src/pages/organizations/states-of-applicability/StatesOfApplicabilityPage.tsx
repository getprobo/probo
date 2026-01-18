import {
    Button,
    IconPlusLarge,
    PageHeader,
    Card,
    Thead,
    Tbody,
    Tr,
    Th,
    Td,
    ActionDropdown,
    DropdownItem,
    IconTrashCan,
    Table,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import {
    graphql,
    usePreloadedQuery,
    useFragment,
    usePaginationFragment,
    type PreloadedQuery,
} from "react-relay";
import { usePageTitle } from "@probo/hooks";
import { CreateStateOfApplicabilityDialog } from "./_components/CreateStateOfApplicabilityDialog";
import { useEffect } from "react";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { formatDate } from "@probo/helpers";
import { useParams } from "react-router";
import { SnapshotBanner } from "/components/SnapshotBanner";
import { useDeleteStateOfApplicability } from "./_components/useDeleteStateOfApplicability";
import type { StatesOfApplicabilityPageQuery } from "/__generated__/core/StatesOfApplicabilityPageQuery.graphql";
import type { StatesOfApplicabilityPageFragment$key } from "/__generated__/core/StatesOfApplicabilityPageFragment.graphql";
import type { StatesOfApplicabilityPageListQuery } from "/__generated__/core/StatesOfApplicabilityPageListQuery.graphql";
import type { StatesOfApplicabilityPageRowFragment$key } from "/__generated__/core/StatesOfApplicabilityPageRowFragment.graphql";

export const statesOfApplicabilityPageQuery = graphql`
    query StatesOfApplicabilityPageQuery($organizationId: ID!) {
        organization: node(id: $organizationId) {
            __typename
            ... on Organization {
                id
                canCreateStateOfApplicability: permission(
                    action: "core:state-of-applicability:create"
                )
                ...StatesOfApplicabilityPageFragment
            }
        }
    }
`;

const statesOfApplicabilityPageFragment = graphql`
    fragment StatesOfApplicabilityPageFragment on Organization
    @refetchable(queryName: "StatesOfApplicabilityPageListQuery")
    @argumentDefinitions(
        first: { type: "Int", defaultValue: 50 }
        order: {
            type: "StateOfApplicabilityOrder"
            defaultValue: { direction: DESC, field: CREATED_AT }
        }
        after: { type: "CursorKey", defaultValue: null }
        before: { type: "CursorKey", defaultValue: null }
        last: { type: "Int", defaultValue: null }
        filter: {
            type: "StateOfApplicabilityFilter"
            defaultValue: { snapshotId: null }
        }
    ) {
        statesOfApplicability(
            first: $first
            after: $after
            last: $last
            before: $before
            orderBy: $order
            filter: $filter
        ) @connection(key: "StatesOfApplicabilityPage_statesOfApplicability") {
            __id
            edges {
                node {
                    id
                    ...StatesOfApplicabilityPageRowFragment
                    canDelete: permission(
                        action: "core:state-of-applicability:delete"
                    )
                }
            }
        }
    }
`;

const statesOfApplicabilityPageRowFragment = graphql`
    fragment StatesOfApplicabilityPageRowFragment on StateOfApplicability {
        id
        name
        createdAt
        applicabilityStatementsInfo: applicabilityStatements(first: 0) {
            totalCount
        }
    }
`;

export default function StatesOfApplicabilityPage({
    queryRef,
}: {
    queryRef: PreloadedQuery<StatesOfApplicabilityPageQuery>;
}) {
    const { __ } = useTranslate();
    const { snapshotId } = useParams<{ snapshotId?: string }>();
    const isSnapshotMode = Boolean(snapshotId);

    usePageTitle(__("States of Applicability"));

    const { organization } = usePreloadedQuery(
        statesOfApplicabilityPageQuery,
        queryRef,
    );

    if (organization.__typename !== "Organization") {
        throw new Error("Organization not found");
    }

    const {
        data: { statesOfApplicability },
        loadNext,
        hasNext,
        refetch,
        isLoadingNext,
    } = usePaginationFragment<
        StatesOfApplicabilityPageListQuery,
        StatesOfApplicabilityPageFragment$key
    >(statesOfApplicabilityPageFragment, organization);

    useEffect(() => {
        if (snapshotId) {
            refetch(
                { filter: { snapshotId } },
                { fetchPolicy: "store-or-network" },
            );
        }
    }, [snapshotId, refetch]);

    const hasAnyAction =
        !isSnapshotMode &&
        statesOfApplicability.edges.some(({ node }) => node.canDelete);

    return (
        <div className="space-y-6">
            {snapshotId && <SnapshotBanner snapshotId={snapshotId} />}
            <PageHeader
                title={__("States of Applicability")}
                description={__(
                    "Manage states of applicability for your organization's frameworks.",
                )}
            >
                {!isSnapshotMode &&
                    organization.canCreateStateOfApplicability && (
                        <CreateStateOfApplicabilityDialog
                            connectionId={statesOfApplicability.__id}
                        >
                            <Button icon={IconPlusLarge}>
                                {__("Add state of applicability")}
                            </Button>
                        </CreateStateOfApplicabilityDialog>
                    )}
            </PageHeader>

            {statesOfApplicability.edges.length > 0 ? (
                <Card>
                    <Table>
                        <Thead>
                            <Tr>
                                <Th>{__("Name")}</Th>
                                <Th>{__("Created at")}</Th>
                                <Th>{__("Statements")}</Th>
                                {hasAnyAction && <Th>{__("Actions")}</Th>}
                            </Tr>
                        </Thead>
                        <Tbody>
                            {statesOfApplicability.edges.map(({ node }) => (
                                <StateOfApplicabilityRow
                                    key={node.id}
                                    fKey={node}
                                    canDelete={node.canDelete}
                                    connectionId={statesOfApplicability.__id}
                                    hasAnyAction={hasAnyAction}
                                />
                            ))}
                        </Tbody>
                    </Table>

                    {hasNext && (
                        <div className="p-4 border-t">
                            <Button
                                variant="secondary"
                                onClick={() => loadNext(50)}
                                disabled={isLoadingNext}
                            >
                                {isLoadingNext
                                    ? __("Loading...")
                                    : __("Load more")}
                            </Button>
                        </div>
                    )}
                </Card>
            ) : (
                <Card padded>
                    <div className="text-center py-12">
                        <h3 className="text-lg font-semibold mb-2">
                            {__("No states of applicability yet")}
                        </h3>
                        <p className="text-txt-tertiary mb-4">
                            {__(
                                "Create your first state of applicability to get started.",
                            )}
                        </p>
                    </div>
                </Card>
            )}
        </div>
    );
}

function StateOfApplicabilityRow({
    fKey,
    canDelete,
    connectionId,
    hasAnyAction,
}: {
    fKey: StatesOfApplicabilityPageRowFragment$key;
    canDelete: boolean;
    connectionId: string;
    hasAnyAction: boolean;
}) {
    const { __ } = useTranslate();
    const organizationId = useOrganizationId();
    const { snapshotId } = useParams<{ snapshotId?: string }>();

    const stateOfApplicability = useFragment(
        statesOfApplicabilityPageRowFragment,
        fKey,
    );

    const deleteStateOfApplicability = useDeleteStateOfApplicability(
        stateOfApplicability,
        connectionId,
    );

    const detailUrl = snapshotId
        ? `/organizations/${organizationId}/snapshots/${snapshotId}/states-of-applicability/${stateOfApplicability.id}`
        : `/organizations/${organizationId}/states-of-applicability/${stateOfApplicability.id}`;

    return (
        <Tr to={detailUrl}>
            <Td>{stateOfApplicability.name}</Td>
            <Td>
                <time dateTime={stateOfApplicability.createdAt}>
                    {formatDate(stateOfApplicability.createdAt)}
                </time>
            </Td>
            <Td>
                {stateOfApplicability.applicabilityStatementsInfo?.totalCount ??
                    0}
            </Td>
            {hasAnyAction && (
                <Td noLink width={50} className="text-end">
                    <ActionDropdown>
                        {canDelete && (
                            <DropdownItem
                                icon={IconTrashCan}
                                variant="danger"
                                onSelect={(e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    deleteStateOfApplicability();
                                }}
                            >
                                {__("Delete")}
                            </DropdownItem>
                        )}
                    </ActionDropdown>
                </Td>
            )}
        </Tr>
    );
}
