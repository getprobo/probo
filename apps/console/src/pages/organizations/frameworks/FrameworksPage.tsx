import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Card,
  DropdownItem,
  FileButton,
  FrameworkLogo,
  FrameworkSelector,
  IconFolderUpload,
  IconPencil,
  IconTrashCan,
  PageHeader,
  useDialogRef,
} from "@probo/ui";
import { type ChangeEventHandler, useState } from "react";
import {
  graphql,
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link } from "react-router";

import type { FrameworkGraphListQuery } from "#/__generated__/core/FrameworkGraphListQuery.graphql";
import type { FrameworksPageCardFragment$key } from "#/__generated__/core/FrameworksPageCardFragment.graphql";
import {
  frameworksQuery,
  useDeleteFrameworkMutation,
} from "#/hooks/graph/FrameworkGraph";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { FrameworkFormDialog } from "./dialogs/FrameworkFormDialog";

type Props = {
  queryRef: PreloadedQuery<FrameworkGraphListQuery>;
};

const importFrameworkMutation = graphql`
  mutation FrameworksPageImportMutation(
    $input: ImportFrameworkInput!
    $connections: [ID!]!
  ) {
    importFramework(input: $input) {
      frameworkEdge @prependEdge(connections: $connections) {
        node {
          id
          ...FrameworksPageCardFragment
        }
      }
    }
  }
`;

export default function FrameworksPage(props: Props) {
  const { __ } = useTranslate();
  usePageTitle(__("Frameworks"));
  const data = usePreloadedQuery(frameworksQuery, props.queryRef);
  const connectionId = data.organization.frameworks!.__id;
  const frameworks
    = data.organization.frameworks?.edges.map(edge => edge.node) ?? [];
  const [commitImport, isImporting] = useMutationWithToasts(
    importFrameworkMutation,
    {
      successMessage: __("Framework imported successfully"),
      errorMessage: __("Failed to import framework"),
    },
  );
  const [isUploading, setUploading] = useState(false);
  const dialogRef = useDialogRef();

  const importNamedFramework = async (name: string) => {
    // For custom framework, open the form
    if (name === "custom") {
      console.log(name, dialogRef);
      dialogRef.current?.open();
      return;
    }
    // Otherwise load the JSON and send the file to the server
    try {
      setUploading(true);
      const fileName = `${name}.json`;
      const json = await fetch(`/data/frameworks/${fileName}`).then(res =>
        res.text(),
      );
      const file = new File([json], fileName, {
        type: "application/json",
      });
      await importFile(file);
    } finally {
      setUploading(false);
    }
  };

  const importFile = (file: File) => {
    return commitImport({
      variables: {
        input: {
          organizationId: data.organization.id!,
          file: null,
        },
        connections: [connectionId],
      },
      uploadables: {
        "input.file": file,
      },
    });
  };

  const handleUpload: ChangeEventHandler<HTMLInputElement> = (event) => {
    const input = event.currentTarget;
    const file = input.files?.[0];
    if (!file) {
      return;
    }
    void importFile(file).finally(() => {
      input.value = "";
    });
  };

  const isLoading = isUploading || isImporting;

  const hasAnyAction = frameworks.some(
    ({ canUpdate, canDelete }) => canUpdate || canDelete,
  );

  return (
    <div className="space-y-6">
      <FrameworkFormDialog
        ref={dialogRef}
        connectionId={connectionId}
        organizationId={data.organization.id!}
      />
      <PageHeader
        title={__("Frameworks")}
        description={__("Manage your compliance frameworks")}
      >
        {data.organization.canCreateFramework && (
          <>
            <FileButton
              variant="secondary"
              icon={IconFolderUpload}
              onChange={handleUpload}
              disabled={isLoading}
            >
              {__("Import")}
            </FileButton>
            <FrameworkSelector
              onSelect={name => void importNamedFramework(name)}
              disabled={isLoading}
            />
          </>
        )}
      </PageHeader>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {frameworks.map(framework => (
          <FrameworkCard
            organizationId={data.organization.id!}
            connectionId={connectionId}
            key={framework.id}
            framework={framework}
            hasAnyAction={hasAnyAction}
          />
        ))}
      </div>
    </div>
  );
}

const frameworkCardFragment = graphql`
  fragment FrameworksPageCardFragment on Framework {
    id
    name
    description
    lightLogoURL
    darkLogoURL
    canUpdate: permission(action: "core:framework:update")
    canDelete: permission(action: "core:framework:delete")
  }
`;

type FrameworkCardProps = {
  organizationId: string;
  connectionId: string;
  framework: FrameworksPageCardFragment$key;
  hasAnyAction: boolean;
};

function FrameworkCard(props: FrameworkCardProps) {
  const framework = useFragment(frameworkCardFragment, props.framework);
  const deleteFramework = useDeleteFrameworkMutation(
    framework,
    props.connectionId,
  );
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  return (
    <Card padded className="p-6 bg-white rounded shadow relative">
      <FrameworkFormDialog
        ref={dialogRef}
        connectionId={props.connectionId}
        organizationId={props.organizationId}
        framework={framework}
      />
      <div className="flex justify-between mb-3">
        <FrameworkLogo
          name={framework.name}
          lightLogoURL={framework.lightLogoURL}
          darkLogoURL={framework.darkLogoURL}
        />
        {props.hasAnyAction && (
          <ActionDropdown className="z-10 relative">
            {framework.canUpdate && (
              <DropdownItem
                icon={IconPencil}
                onClick={() => {
                  dialogRef.current?.open();
                }}
              >
                {__("Edit")}
              </DropdownItem>
            )}
            {framework.canDelete && (
              <DropdownItem
                icon={IconTrashCan}
                onClick={() => deleteFramework()}
                variant="danger"
              >
                {__("Delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        )}
      </div>
      <h2 className="text-xl font-medium">
        <Link
          className="hover:underline after:absolute after:content-[''] after:inset-0"
          to={`/organizations/${props.organizationId}/frameworks/${framework.id}`}
        >
          {framework.name}
        </Link>
      </h2>
      <p className="text-sm text-txt-secondary">{framework.description}</p>
    </Card>
  );
}
