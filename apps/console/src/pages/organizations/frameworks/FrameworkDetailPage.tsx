import {
    useFragment,
    usePreloadedQuery,
    type PreloadedQuery,
} from "react-relay";
import { usePageTitle } from "@probo/hooks";
import { ConnectionHandler, graphql } from "relay-runtime";
import {
    ActionDropdown,
    Button,
    ControlItem,
    DropdownItem,
    FrameworkLogo,
    IconPencil,
    IconPlusLarge,
    IconTrashCan,
    PageHeader,
} from "@probo/ui";
import {
    connectionListKey,
    frameworkNodeQuery,
    useDeleteFrameworkMutation,
} from "/hooks/graph/FrameworkGraph";
import { useTranslate } from "@probo/i18n";
import { Navigate, Outlet, useNavigate, useParams } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { FrameworkGraphNodeQuery } from "/__generated__/core/FrameworkGraphNodeQuery.graphql";
import type { FrameworkDetailPageFragment$key } from "/__generated__/core/FrameworkDetailPageFragment.graphql";
import type { FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation } from "/__generated__/core/FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation.graphql";
import type { FrameworkDetailPageExportFrameworkMutation } from "/__generated__/core/FrameworkDetailPageExportFrameworkMutation.graphql";
import { FrameworkFormDialog } from "./dialogs/FrameworkFormDialog";
import { FrameworkControlDialog } from "./dialogs/FrameworkControlDialog";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";

const frameworkDetailFragment = graphql`
    fragment FrameworkDetailPageFragment on Framework {
        id
        name
        description
        lightLogoURL
        darkLogoURL
        canExport: permission(action: "core:franework:export")
        canUpdate: permission(action: "core:framework:update")
        canDelete: permission(action: "core:framework:delete")
        canCreateControl: permission(action: "core:control:create")
        organization {
            name
        }
        controls(
            first: 250
            orderBy: { field: SECTION_TITLE, direction: ASC }
        ) {
            __id
            edges {
                node {
                    id
                    sectionTitle
                    name
                    status
                    exclusionJustification
                    bestPractice
                }
            }
        }
    }
`;

const generateFrameworkStateOfApplicabilityMutation = graphql`
    mutation FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation(
        $frameworkId: ID!
    ) {
        generateFrameworkStateOfApplicability(
            input: { frameworkId: $frameworkId }
        ) {
            data
        }
    }
`;

const exportFrameworkMutation = graphql`
    mutation FrameworkDetailPageExportFrameworkMutation($frameworkId: ID!) {
        exportFramework(input: { frameworkId: $frameworkId }) {
            exportJobId
        }
    }
`;

type Props = {
    queryRef: PreloadedQuery<FrameworkGraphNodeQuery>;
};

export default function FrameworkDetailPage(props: Props) {
    const { queryRef } = props;

    const { __ } = useTranslate();
    const { controlId } = useParams<{ controlId?: string }>();
    const organizationId = useOrganizationId();
    const data = usePreloadedQuery<FrameworkGraphNodeQuery>(
        frameworkNodeQuery,
        queryRef,
    );
    const framework = useFragment<FrameworkDetailPageFragment$key>(
        frameworkDetailFragment,
        data.node,
    );
    const navigate = useNavigate();
    const controls = framework.controls.edges.map((edge) => edge.node);
    const selectedControl = controlId
        ? controls.find((control) => control.id === controlId)
        : controls[0] || null;
    const connectionId = framework.controls.__id;
    const deleteFramework = useDeleteFrameworkMutation(
        framework,
        ConnectionHandler.getConnectionID(organizationId, connectionListKey)!,
    );
    const [generateFrameworkStateOfApplicability] =
        useMutationWithToasts<FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation>(
            generateFrameworkStateOfApplicabilityMutation,
            {
                errorMessage:
                    "Failed to generate framework state of applicability",
                successMessage:
                    "Framework state of applicability generated successfully",
            },
        );

    const [exportFramework] =
        useMutationWithToasts<FrameworkDetailPageExportFrameworkMutation>(
            exportFrameworkMutation,
            {
                errorMessage: "Failed to export framework",
                successMessage:
                    "Framework export started successfully. You will receive an email when the export is ready.",
            },
        );

    usePageTitle(`${framework.name} | ${selectedControl?.sectionTitle}`);
    const onDelete = () => {
        deleteFramework({
            onSuccess: () => {
                navigate(`/organizations/${organizationId}/frameworks`);
            },
        });
    };

    if (!controlId && controls.length > 0) {
        return (
            <Navigate
                to={`/organizations/${organizationId}/frameworks/${framework.id}/controls/${controls[0].id}`}
            />
        );
    }

    return (
        <div className="space-y-6">
            <PageHeader
                title={
                    <>
                        <FrameworkLogo {...framework} />
                        {framework.name}
                    </>
                }
            >
                {framework.canUpdate && (
                    <FrameworkFormDialog
                        organizationId={organizationId}
                        framework={framework}
                    >
                        <Button icon={IconPencil} variant="secondary">
                            {__("Edit")}
                        </Button>
                    </FrameworkFormDialog>
                )}
                <ActionDropdown variant="secondary">
                    <DropdownItem
                        variant="primary"
                        onClick={() => {
                            generateFrameworkStateOfApplicability({
                                variables: { frameworkId: framework.id },
                                onCompleted: (data) => {
                                    if (
                                        data
                                            .generateFrameworkStateOfApplicability
                                            ?.data
                                    ) {
                                        const link =
                                            window.document.createElement("a");
                                        link.href =
                                            data.generateFrameworkStateOfApplicability.data;
                                        link.download = `${framework.organization.name}-${framework.name}-SOA.xlsx`;
                                        window.document.body.appendChild(link);
                                        link.click();
                                        window.document.body.removeChild(link);
                                    }
                                },
                            });
                        }}
                    >
                        {__("Download SOA")}
                    </DropdownItem>
                    <DropdownItem
                        variant="primary"
                        onClick={() => {
                            exportFramework({
                                variables: { frameworkId: framework.id },
                            });
                        }}
                    >
                        {__("Export Framework")}
                    </DropdownItem>
                    {framework.canDelete && (
                        <DropdownItem
                            icon={IconTrashCan}
                            variant="danger"
                            onClick={onDelete}
                        >
                            {__("Delete")}
                        </DropdownItem>
                    )}
                </ActionDropdown>
            </PageHeader>
            <div className="text-lg font-semibold">
                {__("Requirement categories")}
            </div>
            <div className="divide-x divide-border-low grid grid-cols-[264px_1fr]">
                <div
                    className="space-y-1 overflow-y-auto pr-6 mr-6 sticky top-0"
                    style={{ maxHeight: "calc(100vh - 48px)" }}
                >
                    {controls.map((control) => (
                        <ControlItem
                            key={control.id}
                            id={control.sectionTitle}
                            description={control.name}
                            excluded={control.status === "EXCLUDED"}
                            to={`/organizations/${organizationId}/frameworks/${framework.id}/controls/${control.id}`}
                            active={selectedControl?.id === control.id}
                        />
                    ))}
                    {framework.canCreateControl && (
                        <FrameworkControlDialog
                            frameworkId={framework.id}
                            connectionId={connectionId}
                        >
                            <button className="flex gap-[6px] flex-col w-full p-4 space-y-[6px] rounded-xl cursor-pointer text-start text-sm text-txt-tertiary hover:bg-tertiary-hover">
                                <IconPlusLarge
                                    size={20}
                                    className="text-txt-primary"
                                />
                                {__("Add new control")}
                            </button>
                        </FrameworkControlDialog>
                    )}
                </div>
                <Outlet context={{ framework }} />
            </div>
        </div>
    );
}
