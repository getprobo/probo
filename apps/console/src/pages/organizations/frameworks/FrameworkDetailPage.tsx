import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
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
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import { Navigate, Outlet, useNavigate, useParams } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { FrameworkDetailPageExportFrameworkMutation } from "#/__generated__/core/FrameworkDetailPageExportFrameworkMutation.graphql";
import type { FrameworkDetailPageFragment$key } from "#/__generated__/core/FrameworkDetailPageFragment.graphql";
import type { FrameworkGraphNodeQuery } from "#/__generated__/core/FrameworkGraphNodeQuery.graphql";
import {
  connectionListKey,
  frameworkNodeQuery,
  useDeleteFrameworkMutation,
} from "#/hooks/graph/FrameworkGraph";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { FrameworkControlDialog } from "./dialogs/FrameworkControlDialog";
import { FrameworkFormDialog } from "./dialogs/FrameworkFormDialog";

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
                    bestPractice
                }
            }
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
  const controls = framework.controls.edges.map(edge => edge.node);
  const selectedControl = controlId
    ? controls.find(control => control.id === controlId)
    : controls[0] || null;
  const connectionId = framework.controls.__id;
  const deleteFramework = useDeleteFrameworkMutation(
    framework,
    ConnectionHandler.getConnectionID(organizationId, connectionListKey),
  );
  const [exportFramework]
    = useMutationWithToasts<FrameworkDetailPageExportFrameworkMutation>(
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
        void navigate(`/organizations/${organizationId}/frameworks`);
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

  const hasAnyAction = framework.canExport || framework.canDelete;

  return (
    <div className="space-y-6">
      <PageHeader
        title={(
          <>
            <FrameworkLogo {...framework} />
            {framework.name}
          </>
        )}
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
        {hasAnyAction && (
          <ActionDropdown variant="secondary">
            {framework.canExport
              && (
                <DropdownItem
                  variant="primary"
                  onClick={() => {
                    void exportFramework({
                      variables: { frameworkId: framework.id },
                    });
                  }}
                >
                  {__("Export Framework")}
                </DropdownItem>
              )}
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
        )}
      </PageHeader>
      <div className="text-lg font-semibold">
        {__("Requirement categories")}
      </div>
      <div className="divide-x divide-border-low grid grid-cols-[264px_1fr]">
        <div
          className="space-y-1 overflow-y-auto pr-6 mr-6 sticky top-0"
          style={{ maxHeight: "calc(100vh - 48px)" }}
        >
          {controls.map(control => (
            <ControlItem
              key={control.id}
              id={control.sectionTitle}
              description={control.name}
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
