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
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
} from "@probo/ui";
import { FrameworkLogo } from "/components/FrameworkLogo";
import {
  connectionListKey,
  frameworkNodeQuery,
  useDeleteFrameworkMutation,
} from "/hooks/graph/FrameworkGraph";
import { useTranslate } from "@probo/i18n";
import { Navigate, Outlet, useNavigate, useParams } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { FrameworkGraphNodeQuery } from "/hooks/graph/__generated__/FrameworkGraphNodeQuery.graphql";
import type { FrameworkDetailPageFragment$key } from "./__generated__/FrameworkDetailPageFragment.graphql";
import type { FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation } from "./__generated__/FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation.graphql";
import { FrameworkFormDialog } from "./dialogs/FrameworkFormDialog";
import { FrameworkControlDialog } from "./dialogs/FrameworkControlDialog";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";

const frameworkDetailFragment = graphql`
  fragment FrameworkDetailPageFragment on Framework {
    id
    name
    description
    organization {
      name
    }
    controls(first: 250, orderBy: { field: SECTION_TITLE, direction: ASC }) {
      __id
      edges {
        node {
          id
          sectionTitle
          name
          description
          status
          exclusionJustification
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

type Props = {
  queryRef: PreloadedQuery<FrameworkGraphNodeQuery>;
};

export default function FrameworkDetailPage(props: Props) {
  const { __ } = useTranslate();
  const { controlId } = useParams<{ controlId?: string }>();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery(frameworkNodeQuery, props.queryRef);
  const framework = useFragment<FrameworkDetailPageFragment$key>(
    frameworkDetailFragment,
    data.node
  );
  const navigate = useNavigate();
  const controls = framework.controls.edges.map((edge) => edge.node);
  const selectedControl = controlId
    ? controls.find((control) => control.id === controlId)
    : controls[0] || null;
  const connectionId = framework.controls.__id;
  const deleteFramework = useDeleteFrameworkMutation(
    framework,
    ConnectionHandler.getConnectionID(organizationId, connectionListKey)!
  );

  const [generateFrameworkStateOfApplicability] =
    useMutationWithToasts<FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation>(
      generateFrameworkStateOfApplicabilityMutation,
      {
        errorMessage: "Failed to generate framework state of applicability",
        successMessage:
          "Framework state of applicability generated successfully",
      }
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
        <FrameworkFormDialog
          organizationId={organizationId}
          framework={framework}
        >
          <Button icon={IconPencil} variant="secondary">
            {__("Edit")}
          </Button>
        </FrameworkFormDialog>
        <ActionDropdown variant="secondary">
          <DropdownItem
            variant="primary"
            onClick={() => {
              generateFrameworkStateOfApplicability({
                variables: { frameworkId: framework.id },
                onCompleted: (data) => {
                  if (data.generateFrameworkStateOfApplicability?.data) {
                    const link = window.document.createElement("a");
                    link.href = data.generateFrameworkStateOfApplicability.data;
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
          <DropdownItem icon={IconTrashCan} variant="danger" onClick={onDelete}>
            {__("Delete")}
          </DropdownItem>
        </ActionDropdown>
      </PageHeader>
      {framework.description && (
        <div className="bg-secondary border border-border-low rounded-lg p-4 mb-6">
          <div className="text-sm text-txt-secondary">
            {framework.description}
          </div>
        </div>
      )}
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
          <FrameworkControlDialog
            frameworkId={framework.id}
            connectionId={connectionId}
          >
            <button className="flex gap-[6px] flex-col w-full p-4 space-y-[6px] rounded-xl cursor-pointer text-start text-sm text-txt-tertiary hover:bg-tertiary-hover">
              <IconPlusLarge size={20} className="text-txt-primary" />
              {__("Add new control")}
            </button>
          </FrameworkControlDialog>
        </div>
        <Outlet context={{ framework }} />
      </div>
    </div>
  );
}
