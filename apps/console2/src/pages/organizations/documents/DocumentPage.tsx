import type { PreloadedQuery } from "react-relay";
import type { DocumentGraphNodeQuery } from "/hooks/graph/__generated__/DocumentGraphNodeQuery.graphql";
import {
  ConnectionHandler,
  graphql,
  loadQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import {
  DocumentsConnectionKey,
  documentNodeQuery,
  useDeleteDocumentMutation,
} from "/hooks/graph/DocumentGraph";
import { usePageTitle } from "@probo/hooks";
import type { DocumentPageDocumentFragment$key } from "./__generated__/DocumentPageDocumentFragment.graphql";
import { useTranslate } from "@probo/i18n";
import {
  PageHeader,
  Breadcrumb,
  IconCheckmark1,
  Markdown,
  PropertyRow,
  Drawer,
  Badge,
  Avatar,
  DropdownItem,
  ActionDropdown,
  IconTrashCan,
  IconPencil,
  IconClock,
  IconSignature,
  useConfirm,
} from "@probo/ui";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { Button } from "@probo/ui";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { sprintf } from "@probo/helpers";
import { useNavigate } from "react-router";
import UpdateVersionDialog from "./dialogs/UpdateVersionDialog";
import { useRef } from "react";
import { DocumentVersionHistoryDialog } from "./dialogs/DocumentVersionHistoryDialog";
import { DocumentSignaturesDialog } from "./dialogs/DocumentSignaturesDialog";

type Props = {
  queryRef: PreloadedQuery<DocumentGraphNodeQuery>;
};

const documentFragment = graphql`
  fragment DocumentPageDocumentFragment on Document {
    id
    title
    owner {
      id
      fullName
    }
    versions(first: 20) @connection(key: "DocumentPage_versions") {
      __id
      edges {
        node {
          id
          content
          status
          publishedAt
          version
          updatedAt
          signatures(first: 100) @connection(key: "DocumentPage_signatures") {
            __id
            edges {
              node {
                id
                state
                signedBy {
                  id
                }
                ...DocumentSignaturesDialog_signature
              }
            }
          }
          ...DocumentVersionHistoryDialogFragment
          ...DocumentSignaturesDialog_version
        }
      }
    }
  }
`;

const publishDocumentVersionMutation = graphql`
  mutation DocumentPagePublishMutation($input: PublishDocumentVersionInput!) {
    publishDocumentVersion(input: $input) {
      document {
        id
      }
    }
  }
`;

export default function DocumentPage(props: Props) {
  const node = usePreloadedQuery(documentNodeQuery, props.queryRef).node;
  const document = useFragment(
    documentFragment,
    node as DocumentPageDocumentFragment$key
  );
  const { __, dateFormat } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const lastVersion = document.versions.edges[0].node;
  const isDraft = lastVersion.status === "DRAFT";
  const [publishDocumentVersion, isPublishing] = useMutationWithToasts(
    publishDocumentVersionMutation,
    {
      successMessage: __("Document published successfully."),
      errorMessage: __("Failed to publish document. Please try again."),
    }
  );
  const [deleteDocument, isDeleting] = useDeleteDocumentMutation();
  const versionConnectionId = document.versions.__id;

  usePageTitle(document.title);

  const handlePublish = () => {
    publishDocumentVersion({
      variables: {
        input: { documentId: document.id },
      },
      onSuccess: () => {
        // Refresh the whole query to get the new version
        loadQuery(
          props.queryRef.environment,
          documentNodeQuery,
          props.queryRef.variables,
          { fetchPolicy: "network-only" }
        );
      },
    });
  };

  const confirm = useConfirm();

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          const connectionId = ConnectionHandler.getConnectionID(
            organizationId,
            DocumentsConnectionKey
          );
          deleteDocument({
            variables: {
              input: { documentId: document.id },
              connections: [connectionId],
            },
            onSuccess() {
              navigate(`/organizations/${organizationId}/documents`);
            },
            onError: () => resolve(),
          });
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

  const updateDialogRef = useRef<{ open: () => void }>(null);

  return (
    <>
      <UpdateVersionDialog
        ref={updateDialogRef}
        document={document}
        connectionId={versionConnectionId}
      />
      <div className="space-y-6">
        <div className="flex justify-between items-center mb-4">
          <Breadcrumb
            items={[
              {
                label: __("Documents"),
                to: `/organizations/${organizationId}/documents`,
              },
              {
                label: document.title,
              },
            ]}
          />
          <div className="flex gap-2">
            {isDraft && (
              <Button
                onClick={handlePublish}
                icon={IconCheckmark1}
                disabled={isPublishing}
              >
                {__("Publish")}
              </Button>
            )}
            <DocumentVersionHistoryDialog document={document}>
              <Button icon={IconClock} variant="secondary">
                {__("Version history")}
              </Button>
            </DocumentVersionHistoryDialog>
            <DocumentSignaturesDialog document={document}>
              <Button icon={IconSignature} variant="secondary">
                {__("Signatures history")}
              </Button>
            </DocumentSignaturesDialog>

            <ActionDropdown variant="secondary">
              <DropdownItem
                onClick={() => updateDialogRef.current?.open()}
                icon={IconPencil}
              >
                {isDraft ? __("Edit draft document") : __("Create new draft")}
              </DropdownItem>
              <DropdownItem
                variant="danger"
                icon={IconTrashCan}
                disabled={isDeleting}
                onClick={handleDelete}
              >
                {__("Delete")}
              </DropdownItem>
            </ActionDropdown>
          </div>
        </div>
        <PageHeader title={document.title} />
        <Markdown content={lastVersion.content} />
      </div>
      <Drawer>
        <div className="text-base text-txt-primary font-medium mb-4">
          {__("Properties")}
        </div>
        <PropertyRow label={__("Owner")}>
          <Badge variant="highlight" size="md" className="gap-2">
            <Avatar name={document.owner?.fullName ?? ""} />
            {document.owner?.fullName}
          </Badge>
        </PropertyRow>
        <PropertyRow label={__("Status")}>
          <Badge
            variant={isDraft ? "highlight" : "success"}
            size="md"
            className="gap-2"
          >
            {isDraft ? __("Draft") : __("Published")}
          </Badge>
        </PropertyRow>
        <PropertyRow label={__("Version")}>
          <div className="text-sm text-txt-secondary">
            {lastVersion.version}
          </div>
        </PropertyRow>
        <PropertyRow label={__("Last modified")}>
          <div className="text-sm text-txt-secondary">
            {dateFormat(lastVersion.updatedAt)}
          </div>
        </PropertyRow>
        {lastVersion.publishedAt && (
          <PropertyRow label={__("Published Date")}>
            <div className="text-sm text-txt-secondary">
              {dateFormat(lastVersion.publishedAt)}
            </div>
          </PropertyRow>
        )}
      </Drawer>
    </>
  );
}
