import { useTranslate } from "@probo/i18n";
import {
  PageHeader,
  Table,
  Tbody,
  Thead,
  Tr,
  Th,
  Td,
  Avatar,
  Badge,
  IconTrashCan,
  Button,
  IconPlusLarge,
  useConfirm,
  ActionDropdown,
  DropdownItem,
} from "@probo/ui";
import {
  useFragment,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";
import type { DocumentGraphListQuery } from "/hooks/graph/__generated__/DocumentGraphListQuery.graphql";
import {
  documentsQuery,
  useDeleteDocumentMutation,
} from "/hooks/graph/DocumentGraph";
import type { DocumentsPageListFragment$key } from "./__generated__/DocumentsPageListFragment.graphql";
import { usePageTitle } from "@probo/hooks";
import { sprintf } from "@probo/helpers";
import { CreateDocumentDialog } from "./dialogs/CreateDocumentDialog";
import type { DocumentsPageRowFragment$key } from "./__generated__/DocumentsPageRowFragment.graphql";

const documentsFragment = graphql`
  fragment DocumentsPageListFragment on Organization {
    documents(first: 100) @connection(key: "DocumentsPageFragment_documents") {
      __id
      edges {
        node {
          id
          ...DocumentsPageRowFragment
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<DocumentGraphListQuery>;
};

export default function DocumentsPage(props: Props) {
  const { __ } = useTranslate();

  const organization = usePreloadedQuery(
    documentsQuery,
    props.queryRef
  ).organization;
  const data = useFragment<DocumentsPageListFragment$key>(
    documentsFragment,
    organization
  );

  const documents = data.documents.edges.map((edge) => edge.node);
  const connectionId = data.documents.__id;
  usePageTitle(__("Documents"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Documents")}
        description={__("Manage your organization’s documents")}
      >
        <CreateDocumentDialog
          connection={connectionId}
          trigger={<Button icon={IconPlusLarge}>{__("New document")}</Button>}
        />
      </PageHeader>
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Status")}</Th>
            <Th>{__("Owner")}</Th>
            <Th>{__("Last update")}</Th>
            <Th>{__("Signatures")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {documents.map((document) => (
            <DocumentRow
              key={document.id}
              document={document}
              organizationId={organization.id}
              connectionId={connectionId}
            />
          ))}
        </Tbody>
      </Table>
    </div>
  );
}

const rowFragment = graphql`
  fragment DocumentsPageRowFragment on Document {
    id
    title
    description
    updatedAt
    owner {
      id
      fullName
    }
    versions(first: 1) {
      edges {
        node {
          id
          status
          signatures(first: 100) {
            edges {
              node {
                id
                state
              }
            }
          }
        }
      }
    }
  }
`;

function DocumentRow({
  document: documentKey,
  organizationId,
  connectionId,
}: {
  document: DocumentsPageRowFragment$key;
  organizationId: string;
  connectionId: string;
}) {
  const document = useFragment<DocumentsPageRowFragment$key>(
    rowFragment,
    documentKey
  );
  const lastVersion = document.versions.edges[0].node;
  const isDraft = lastVersion.status === "DRAFT";
  const { __, dateFormat } = useTranslate();
  const signatures = lastVersion.signatures.edges.map((edge) => edge.node);
  const signedCount = signatures.filter(
    (signature) => signature.state === "SIGNED"
  ).length;
  const [deleteDocument] = useDeleteDocumentMutation();
  const confirm = useConfirm();

  const handleDelete = () => {
    confirm(
      () =>
        deleteDocument({
          variables: {
            input: { documentId: document.id },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete the document "%s". This action cannot be undone.'
          ),
          document.title
        ),
      }
    );
  };

  return (
    <Tr to={`/organizations/${organizationId}/documents/${document.id}`}>
      <Td>
        <div className="flex gap-4 items-center">
          <img
            src="/document.png"
            alt=""
            width={28}
            height={36}
            className="border-4 border-highlight rounded box-content"
          />
          {document.title}
        </div>
      </Td>
      <Td>
        <Badge variant={isDraft ? "neutral" : "success"}>
          {isDraft ? __("Draft") : __("Published")}
        </Badge>
      </Td>
      <Td>
        <div className="flex gap-2 items-center">
          <Avatar name={document.owner.fullName} />
          {document.owner.fullName}
        </div>
      </Td>
      <Td>
        {dateFormat(document.updatedAt, {
          year: "numeric",
          month: "short",
          day: "numeric",
          weekday: "short",
        })}
      </Td>
      <Td>
        {signedCount}/{signatures.length}
      </Td>
      <Td noLink width={50} className="text-end">
        <ActionDropdown>
          <DropdownItem
            variant="danger"
            icon={IconTrashCan}
            onClick={handleDelete}
          >
            {__("Delete")}
          </DropdownItem>
        </ActionDropdown>
      </Td>
    </Tr>
  );
}
