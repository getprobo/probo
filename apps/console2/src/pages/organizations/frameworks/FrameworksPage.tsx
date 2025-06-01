import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Avatar,
  Card,
  DropdownItem,
  FileButton,
  FrameworkSelector,
  IconTrashCan,
  IconUpload,
  PageHeader,
  useDialogRef,
} from "@probo/ui";
import {
  graphql,
  useFragment,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import type { FrameworkGraphListQuery } from "/hooks/graph/__generated__/FrameworkGraphListQuery.graphql";
import {
  frameworksQuery,
  useDeleteFrameworkMutation,
} from "/hooks/graph/FrameworkGraph";
import { Link } from "react-router";
import type { FrameworksPageCardFragment$key } from "./__generated__/FrameworksPageCardFragment.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useState, type ChangeEventHandler } from "react";
import { availableFrameworks } from "@probo/helpers";
import { CreateFrameworkDialog } from "./dialogs/CreateFrameworkDialog";

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

const availableLogos = new Map(
  availableFrameworks.map((framework) => [framework.name, framework.logo])
);

export default function FrameworksPage(props: Props) {
  const { __ } = useTranslate();
  usePageTitle(__("Frameworks"));
  const data = usePreloadedQuery(frameworksQuery, props.queryRef);
  const connectionId = data.organization.frameworks!.__id;
  const frameworks =
    data.organization.frameworks?.edges.map((edge) => edge.node) ?? [];
  const [commitImport, isImporting] = useMutationWithToasts(
    importFrameworkMutation,
    {
      successMessage: __("Framework imported successfully"),
      errorMessage: __("Failed to import framework"),
    }
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
      const json = await fetch(`/data/frameworks/${fileName}`).then((res) =>
        res.text()
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
    importFile(file).finally(() => {
      input.value = "";
    });
  };

  const isLoading = isUploading || isImporting;

  return (
    <div className="space-y-6">
      <CreateFrameworkDialog
        ref={dialogRef}
        connectionId={connectionId}
        organizationId={data.organization.id!}
      />
      <PageHeader
        title={__("Frameworks")}
        description={__("Manage your compliance frameworks")}
      >
        <FileButton
          variant="secondary"
          icon={IconUpload}
          onChange={handleUpload}
          disabled={isLoading}
        >
          {__("Import")}
        </FileButton>
        <FrameworkSelector
          onSelect={importNamedFramework}
          disabled={isLoading}
        />
      </PageHeader>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {frameworks.map((framework) => (
          <FrameworkCard
            organizationId={data.organization.id!}
            connectionId={connectionId}
            key={framework.id}
            framework={framework}
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
  }
`;

type FrameworkCardProps = {
  organizationId: string;
  connectionId: string;
  framework: FrameworksPageCardFragment$key;
};

function FrameworkCard(props: FrameworkCardProps) {
  const framework = useFragment(frameworkCardFragment, props.framework);
  const deleteFramework = useDeleteFrameworkMutation(
    framework,
    props.connectionId
  );
  const { __ } = useTranslate();
  const logo = availableLogos.get(framework.name);
  return (
    <Card padded className="p-6 bg-white rounded shadow relative">
      <div className="flex justify-between mb-3">
        {logo ? (
          <img src={logo} alt="" className="size-12" />
        ) : (
          <Avatar name={framework.name} size="l" className="size-12" />
        )}
        <ActionDropdown className="z-10 relative">
          <DropdownItem
            icon={IconTrashCan}
            onClick={deleteFramework}
            variant="danger"
          >
            {__("Delete")}
          </DropdownItem>
        </ActionDropdown>
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
