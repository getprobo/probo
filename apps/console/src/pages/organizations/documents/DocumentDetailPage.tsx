import type { PreloadedQuery } from "react-relay";
import {
  ConnectionHandler,
  graphql,
  loadQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import type { DocumentGraphNodeQuery } from "/hooks/graph/__generated__/DocumentGraphNodeQuery.graphql";
import {
  documentNodeQuery,
  DocumentsConnectionKey,
  useDeleteDocumentMutation,
} from "/hooks/graph/DocumentGraph";
import { usePageTitle } from "@probo/hooks";
import type {
  DocumentDetailPageDocumentFragment$data,
  DocumentDetailPageDocumentFragment$key,
} from "./__generated__/DocumentDetailPageDocumentFragment.graphql";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Avatar,
  Badge,
  Breadcrumb,
  Button,
  Drawer,
  Dropdown,
  DropdownItem,
  IconCheckmark1,
  IconChevronDown,
  IconClock,
  IconPencil,
  IconTrashCan,
  PageHeader,
  PropertyRow,
  TabBadge,
  TabLink,
  Tabs,
  useConfirm,
} from "@probo/ui";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { getDocumentTypeLabel, sprintf } from "@probo/helpers";
import {
  Link,
  Outlet,
  useLocation,
  useNavigate,
  useParams,
} from "react-router";
import UpdateVersionDialog from "./dialogs/UpdateVersionDialog";
import { useRef } from "react";
import type { NodeOf } from "/types.ts";
import clsx from "clsx";

type Props = {
  queryRef: PreloadedQuery<DocumentGraphNodeQuery>;
};

const documentFragment = graphql`
  fragment DocumentDetailPageDocumentFragment on Document {
    id
    title
    documentType
    owner {
      id
      fullName
    }
    ...DocumentControlsTabFragment
    controlsInfo: controls(first: 0) {
      totalCount
    }
    versions(first: 20) @connection(key: "DocumentDetailPage_versions") {
      __id
      edges {
        node {
          id
          content
          status
          publishedAt
          version
          updatedAt
          publishedBy {
            fullName
          }
          signatures(first: 100)
            @connection(key: "DocumentDetailPage_signatures") {
            __id
            edges {
              node {
                id
                state
                signedBy {
                  id
                }
                ...DocumentSignaturesTab_signature
              }
            }
          }
        }
      }
    }
  }
`;

const publishDocumentVersionMutation = graphql`
  mutation DocumentDetailPagePublishMutation(
    $input: PublishDocumentVersionInput!
  ) {
    publishDocumentVersion(input: $input) {
      document {
        id
      }
    }
  }
`;

export default function DocumentDetailPage(props: Props) {
  const { versionId } = useParams<{ versionId?: string }>();
  const node = usePreloadedQuery(documentNodeQuery, props.queryRef).node;
  const document = useFragment(
    documentFragment,
    node as DocumentDetailPageDocumentFragment$key,
  );
  const { __, dateFormat } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const versions = document.versions.edges.map((edge) => edge.node);
  const currentVersion =
    document.versions.edges.find((v) => v.node.id === versionId)?.node ??
    document.versions.edges[0].node;
  const signatures = currentVersion.signatures.edges.map((s) => s.node);
  const signedSignatures = signatures.filter((s) => s.state === "SIGNED");
  const isDraft = currentVersion.status === "DRAFT";
  const [publishDocumentVersion, isPublishing] = useMutationWithToasts(
    publishDocumentVersionMutation,
    {
      successMessage: __("Document published successfully."),
      errorMessage: __("Failed to publish document. Please try again."),
    },
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
          { fetchPolicy: "network-only" },
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
            DocumentsConnectionKey,
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
            'This will permanently delete the document "%s". This action cannot be undone.',
          ),
          document.title,
        ),
      },
    );
  };

  const updateDialogRef = useRef<{ open: () => void }>(null);
  const controlsCount = document.controlsInfo.totalCount;
  const urlPrefix = versionId
    ? `/organizations/${organizationId}/documents/${document.id}/versions/${versionId}`
    : `/organizations/${organizationId}/documents/${document.id}`;

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
            <Dropdown
              toggle={
                <Button icon={IconClock} variant="secondary">
                  {__("Version history")}
                  <IconChevronDown size={12} />
                </Button>
              }
            >
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
              {versions.map((version) => (
                <DropdownItem asChild key={version.id}>
                  <VersionItem
                    document={document}
                    version={version}
                    active={version.id === currentVersion.id}
                    onSelect={() => {}}
                  />
                </DropdownItem>
              ))}
            </Dropdown>

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

        <Tabs>
          <TabLink to={`${urlPrefix}/description`}>{__("Description")}</TabLink>
          <TabLink to={`${urlPrefix}/controls`}>
            {__("Controls")}
            <TabBadge>{controlsCount}</TabBadge>
          </TabLink>
          <TabLink to={`${urlPrefix}/signatures`}>
            {__("Signatures")}
            <TabBadge>
              {signedSignatures.length}/{signatures.length}
            </TabBadge>
          </TabLink>
        </Tabs>

        <Outlet context={{ document, version: currentVersion }} />
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
        <PropertyRow label={__("Type")}>
          <div className="text-sm text-txt-secondary">
            {getDocumentTypeLabel(__, document.documentType)}
          </div>
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
            {currentVersion.version}
          </div>
        </PropertyRow>
        <PropertyRow label={__("Last modified")}>
          <div className="text-sm text-txt-secondary">
            {dateFormat(currentVersion.updatedAt)}
          </div>
        </PropertyRow>
        {currentVersion.publishedAt && (
          <PropertyRow label={__("Published Date")}>
            <div className="text-sm text-txt-secondary">
              {dateFormat(currentVersion.publishedAt)}
            </div>
          </PropertyRow>
        )}
      </Drawer>
    </>
  );
}

type Version = NodeOf<DocumentDetailPageDocumentFragment$data["versions"]>;

function VersionItem({
  document,
  version,
  active,
  onSelect,
  ...props
}: {
  document: DocumentDetailPageDocumentFragment$data;
  version: Version;
  active?: boolean;
  onSelect: (v: Version) => void;
}) {
  const { dateTimeFormat } = useTranslate();
  const organizationId = useOrganizationId();
  const suffix = useLocation().pathname.split("/").at(-1);
  return (
    <Link
      to={`/organizations/${organizationId}/documents/${document.id}/versions/${version.id}/${suffix}`}
      onClick={() => onSelect(version)}
      className={clsx(
        "flex items-center gap-2 py-2 px-[10px] w-full hover:bg-tertiary-hover cursor-pointer rounded",
        active && "bg-tertiary-pressed",
      )}
      {...props}
    >
      <Avatar
        name={version.publishedBy?.fullName ?? document.owner.fullName}
        size="l"
      />
      <div className="text-start space-y-[2px] w-full overflow-hidden">
        <div className="text-sm text-txt-primary whitespace-nowrap overflow-hidden text-ellipsis">
          {version.publishedBy?.fullName ?? document.owner.fullName}
        </div>
        <div className="text-xs text-txt-secondary whitespace-nowrap overflow-hidden text-ellipsis">
          {dateTimeFormat(version.publishedAt ?? version.updatedAt)}
        </div>
      </div>
    </Link>
  );
}
