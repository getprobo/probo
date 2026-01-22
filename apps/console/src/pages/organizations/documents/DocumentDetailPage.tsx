import {
  documentClassifications,
  documentTypes,
  formatDate,
  getDocumentClassificationLabel,
  getDocumentTypeLabel,
  sprintf,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
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
  IconArrowDown,
  IconCheckmark1,
  IconChevronDown,
  IconClock,
  IconCrossLargeX,
  IconPencil,
  IconTrashCan,
  Input,
  PageHeader,
  PropertyRow,
  TabBadge,
  TabLink,
  Tabs,
  useConfirm,
} from "@probo/ui";
import { clsx } from "clsx";
import { use, useRef, useState } from "react";
import type { PreloadedQuery } from "react-relay";
import {
  graphql,
  loadQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import {
  Link,
  Outlet,
  useLocation,
  useNavigate,
  useParams,
} from "react-router";
import { z } from "zod";

import type {
  DocumentDetailPageDocumentFragment$data,
  DocumentDetailPageDocumentFragment$key,
} from "#/__generated__/core/DocumentDetailPageDocumentFragment.graphql";
import type { DocumentDetailPageExportPDFMutation } from "#/__generated__/core/DocumentDetailPageExportPDFMutation.graphql";
import type { DocumentDetailPageUpdateMutation } from "#/__generated__/core/DocumentDetailPageUpdateMutation.graphql";
import type { DocumentGraphNodeQuery } from "#/__generated__/core/DocumentGraphNodeQuery.graphql";
import {
  PdfDownloadDialog,
  type PdfDownloadDialogRef,
} from "#/components/documents/PdfDownloadDialog";
import { ControlledField } from "#/components/form/ControlledField";
import { DocumentClassificationOptions } from "#/components/form/DocumentClassificationOptions";
import { DocumentTypeOptions } from "#/components/form/DocumentTypeOptions";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import {
  documentNodeQuery,
  useDeleteDocumentMutation,
  useDeleteDraftDocumentVersionMutation,
} from "#/hooks/graph/DocumentGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";
import type { NodeOf } from "#/types";

import UpdateVersionDialog from "./dialogs/UpdateVersionDialog";

type Props = {
  queryRef: PreloadedQuery<DocumentGraphNodeQuery>;
};

const documentFragment = graphql`
  fragment DocumentDetailPageDocumentFragment on Document {
    id
    title
    documentType
    classification
    owner {
      id
      fullName
    }
    canUpdate: permission(action: "core:document:update")
    canDelete: permission(action: "core:document:delete")
    canPublish: permission(action: "core:document-version:publish")
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
          classification
          owner {
            id
            fullName
          }
          canDeleteDraft: permission(
            action: "core:document-version:delete-draft"
          )
          ...DocumentSignaturesTab_version
          signatures(first: 1000)
            @connection(key: "DocumentDetailPage_signatures", filters: []) {
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
        ...DocumentDetailPageDocumentFragment
      }
    }
  }
`;

const exportDocumentVersionPDFMutation = graphql`
  mutation DocumentDetailPageExportPDFMutation(
    $input: ExportDocumentVersionPDFInput!
  ) {
    exportDocumentVersionPDF(input: $input) {
      data
    }
  }
`;

const updateDocumentMutation = graphql`
  mutation DocumentDetailPageUpdateMutation($input: UpdateDocumentInput!) {
    updateDocument(input: $input) {
      document {
        id
        title
        documentType
        classification
        owner {
          id
          fullName
        }
      }
    }
  }
`;

const documentUpdateSchema = z.object({
  title: z.string().min(1, "Title is required").max(255),
  ownerId: z.string().min(1, "Owner is required"),
  documentType: z.enum(documentTypes),
  classification: z.enum(documentClassifications),
});

export default function DocumentDetailPage(props: Props) {
  const { versionId } = useParams<{ versionId?: string }>();
  const node = usePreloadedQuery(documentNodeQuery, props.queryRef).node;
  const document = useFragment<DocumentDetailPageDocumentFragment$key>(
    documentFragment,
    node,
  );
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [isEditingOwner, setIsEditingOwner] = useState(false);
  const [isEditingType, setIsEditingType] = useState(false);
  const [isEditingClassification, setIsEditingClassification] = useState(false);
  const versions = document.versions.edges.map(edge => edge.node);
  const currentVersion
    = document.versions.edges.find(v => v.node.id === versionId)?.node
      ?? document.versions.edges[0].node;
  const signatures = currentVersion.signatures?.edges?.map(s => s.node) ?? [];
  const signedSignatures = signatures.filter(s => s.state === "SIGNED");
  const isDraft = currentVersion.status === "DRAFT";
  const [publishDocumentVersion, isPublishing] = useMutationWithToasts(
    publishDocumentVersionMutation,
    {
      successMessage: __("Document published successfully."),
      errorMessage: __("Failed to publish document"),
    },
  );
  const [deleteDocument, isDeleting] = useDeleteDocumentMutation();
  const [deleteDraftDocumentVersion, isDeletingDraft]
    = useDeleteDraftDocumentVersionMutation();
  const [exportDocumentVersionPDF, isExporting]
    = useMutationWithToasts<DocumentDetailPageExportPDFMutation>(
      exportDocumentVersionPDFMutation,
      {
        successMessage: __("PDF download started."),
        errorMessage: __("Failed to generate PDF"),
      },
    );

  const { email: defaultEmail } = use(CurrentUser);
  const [updateDocument, isUpdatingDocument]
    = useMutationWithToasts<DocumentDetailPageUpdateMutation>(
      updateDocumentMutation,
      {
        successMessage: __("Document updated successfully."),
        errorMessage: __("Failed to update document"),
      },
    );
  const versionConnectionId = document.versions.__id;

  const { register, control, handleSubmit, reset } = useFormWithSchema(
    documentUpdateSchema,
    {
      defaultValues: {
        title: document.title,
        ownerId: currentVersion.owner?.id || "",
        documentType: document.documentType,
        classification: currentVersion.classification,
      },
    },
  );

  usePageTitle(document.title);

  const handleUpdateTitle = async (data: { title: string }) => {
    await updateDocument({
      variables: {
        input: {
          id: document.id,
          title: data.title,
        },
      },
      onSuccess: () => {
        setIsEditingTitle(false);
      },
    });
  };

  const handleUpdateOwner = async (data: { ownerId: string }) => {
    await updateDocument({
      variables: {
        input: {
          id: document.id,
          ownerId: data.ownerId,
        },
      },
      onSuccess: () => {
        setIsEditingOwner(false);
      },
    });
  };

  const handleUpdateDocumentType = async (data: {
    documentType: (typeof documentTypes)[number];
  }) => {
    await updateDocument({
      variables: {
        input: {
          id: document.id,
          documentType: data.documentType,
        },
      },
      onSuccess: () => {
        setIsEditingType(false);
        loadQuery(
          props.queryRef.environment,
          documentNodeQuery,
          props.queryRef.variables,
          { fetchPolicy: "network-only" },
        );
      },
    });
  };

  const handleUpdateClassification = async (data: {
    classification: (typeof documentClassifications)[number];
  }) => {
    await updateDocument({
      variables: {
        input: {
          id: document.id,
          classification: data.classification,
        },
      },
      onSuccess: () => {
        setIsEditingClassification(false);
        loadQuery(
          props.queryRef.environment,
          documentNodeQuery,
          props.queryRef.variables,
          { fetchPolicy: "network-only" },
        );
      },
    });
  };

  const handlePublish = async () => {
    await publishDocumentVersion({
      variables: {
        input: { documentId: document.id },
      },
    });
  };

  const confirm = useConfirm();

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          void deleteDocument({
            variables: {
              input: { documentId: document.id },
            },
            onSuccess() {
              void navigate(`/organizations/${organizationId}/documents`);
              resolve();
            },
            onError: () => resolve(),
          });
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the document \"%s\". This action cannot be undone.",
          ),
          document.title,
        ),
      },
    );
  };

  const handleDeleteDraft = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          void deleteDraftDocumentVersion({
            variables: {
              input: { documentVersionId: currentVersion.id },
              connections: [versionConnectionId],
            },
            onSuccess() {
              loadQuery(
                props.queryRef.environment,
                documentNodeQuery,
                props.queryRef.variables,
                { fetchPolicy: "network-only" },
              );

              resolve();
            },
            onError: () => resolve(),
          });
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the draft version %s of \"%s\". This action cannot be undone.",
          ),
          currentVersion.version,
          document.title,
        ),
      },
    );
  };

  const handleDownloadPdf = async (options: {
    withWatermark: boolean;
    withSignatures: boolean;
    watermarkEmail?: string;
  }) => {
    const input = {
      documentVersionId: currentVersion.id,
      withWatermark: options.withWatermark,
      withSignatures: options.withSignatures,
      ...(options.withWatermark
        && options.watermarkEmail && { watermarkEmail: options.watermarkEmail }),
    };

    await exportDocumentVersionPDF({
      variables: { input },
      onCompleted: (data) => {
        if (data.exportDocumentVersionPDF?.data) {
          const link = window.document.createElement("a");
          link.href = data.exportDocumentVersionPDF.data;
          link.download = `${document.title}-v${currentVersion.version}.pdf`;
          window.document.body.appendChild(link);
          link.click();
          window.document.body.removeChild(link);
        }
      },
    });
  };

  const updateDialogRef = useRef<{ open: () => void }>(null);
  const pdfDownloadDialogRef = useRef<PdfDownloadDialogRef>(null);
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
      <PdfDownloadDialog
        ref={pdfDownloadDialogRef}
        onDownload={options => void handleDownloadPdf(options)}
        isLoading={isExporting}
        defaultEmail={defaultEmail}
      >
        {null}
      </PdfDownloadDialog>
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
            {isDraft && document.canPublish && (
              <Button
                onClick={() => void handlePublish()}
                icon={IconCheckmark1}
                disabled={isPublishing}
              >
                {__("Publish")}
              </Button>
            )}
            <Dropdown
              toggle={(
                <Button icon={IconClock} variant="secondary">
                  {__("Version history")}
                  <IconChevronDown size={12} />
                </Button>
              )}
            >
              {versions.map(version => (
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
              {document.canUpdate && (
                <DropdownItem
                  onClick={() => updateDialogRef.current?.open()}
                  icon={IconPencil}
                >
                  {isDraft ? __("Edit draft document") : __("Create new draft")}
                </DropdownItem>
              )}
              {isDraft
                && versions.length > 1
                && currentVersion.canDeleteDraft && (
                <DropdownItem
                  onClick={handleDeleteDraft}
                  icon={IconTrashCan}
                  disabled={isDeletingDraft}
                >
                  {__("Delete draft document")}
                </DropdownItem>
              )}
              <DropdownItem
                onClick={() => pdfDownloadDialogRef.current?.open()}
                icon={IconArrowDown}
                disabled={isExporting}
              >
                {__("Download PDF")}
              </DropdownItem>
              {document.canDelete && (
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  disabled={isDeleting}
                  onClick={handleDelete}
                >
                  {__("Delete document")}
                </DropdownItem>
              )}
            </ActionDropdown>
          </div>
        </div>
        <PageHeader
          title={
            isEditingTitle
              ? (
                  <div className="flex items-center gap-2">
                    <Input
                      {...register("title")}
                      variant="title"
                      className="flex-1"
                      autoFocus
                      onKeyDown={(e) => {
                        if (e.key === "Escape") {
                          setIsEditingTitle(false);
                          reset();
                        }
                        if (e.key === "Enter") {
                          void handleSubmit(handleUpdateTitle)();
                        }
                      }}
                    />
                    <Button
                      variant="quaternary"
                      icon={IconCheckmark1}
                      onClick={() => void handleSubmit(handleUpdateTitle)()}
                      disabled={isUpdatingDocument}
                    />
                    <Button
                      variant="quaternary"
                      icon={IconCrossLargeX}
                      onClick={() => {
                        setIsEditingTitle(false);
                        reset();
                      }}
                    />
                  </div>
                )
              : (
                  <div className="flex items-center gap-2">
                    <span>{document.title}</span>
                    {document.canUpdate && (
                      <Button
                        variant="quaternary"
                        icon={IconPencil}
                        onClick={() => setIsEditingTitle(true)}
                      />
                    )}
                  </div>
                )
          }
        />

        <Tabs>
          <TabLink to={`${urlPrefix}/description`}>{__("Description")}</TabLink>
          <TabLink to={`${urlPrefix}/controls`}>
            {__("Controls")}
            <TabBadge>{controlsCount}</TabBadge>
          </TabLink>
          {!isDraft && (
            <TabLink to={`${urlPrefix}/signatures`}>
              {__("Signatures")}
              <TabBadge>
                {signedSignatures.length}
                /
                {signatures.length}
              </TabBadge>
            </TabLink>
          )}
        </Tabs>

        <Outlet context={{ document, version: currentVersion }} />
      </div>
      <Drawer>
        <div className="text-base text-txt-primary font-medium mb-4">
          {__("Properties")}
        </div>
        <PropertyRow label={__("Owner")}>
          {isEditingOwner
            ? (
                <EditablePropertyContent
                  onSave={() => void handleSubmit(handleUpdateOwner)()}
                  onCancel={() => {
                    setIsEditingOwner(false);
                    reset();
                  }}
                  disabled={isUpdatingDocument}
                >
                  <PeopleSelectField
                    name="ownerId"
                    control={control}
                    organizationId={organizationId}
                  />
                </EditablePropertyContent>
              )
            : (
                <ReadOnlyPropertyContent
                  onEdit={() => setIsEditingOwner(true)}
                  canEdit={document.canUpdate}
                >
                  <Badge variant="highlight" size="md" className="gap-2">
                    <Avatar name={currentVersion.owner?.fullName ?? ""} />
                    {currentVersion.owner?.fullName}
                  </Badge>
                </ReadOnlyPropertyContent>
              )}
        </PropertyRow>
        <PropertyRow label={__("Type")}>
          {isEditingType
            ? (
                <EditablePropertyContent
                  onSave={() => void handleSubmit(handleUpdateDocumentType)()}
                  onCancel={() => {
                    setIsEditingType(false);
                    reset();
                  }}
                  disabled={isUpdatingDocument}
                >
                  <ControlledField
                    name="documentType"
                    control={control}
                    type="select"
                  >
                    <DocumentTypeOptions />
                  </ControlledField>
                </EditablePropertyContent>
              )
            : (
                <ReadOnlyPropertyContent
                  onEdit={() => setIsEditingType(true)}
                  canEdit={document.canUpdate}
                >
                  <div className="text-sm text-txt-secondary">
                    {getDocumentTypeLabel(__, document.documentType)}
                  </div>
                </ReadOnlyPropertyContent>
              )}
        </PropertyRow>
        <PropertyRow label={__("Classification")}>
          {isEditingClassification
            ? (
                <EditablePropertyContent
                  onSave={() => void handleSubmit(handleUpdateClassification)()}
                  onCancel={() => {
                    setIsEditingClassification(false);
                    reset();
                  }}
                  disabled={isUpdatingDocument}
                >
                  <ControlledField
                    name="classification"
                    control={control}
                    type="select"
                  >
                    <DocumentClassificationOptions />
                  </ControlledField>
                </EditablePropertyContent>
              )
            : (
                <ReadOnlyPropertyContent
                  onEdit={() => setIsEditingClassification(true)}
                  canEdit={document.canUpdate}
                >
                  <div className="text-sm text-txt-secondary">
                    {getDocumentClassificationLabel(
                      __,
                      currentVersion.classification,
                    )}
                  </div>
                </ReadOnlyPropertyContent>
              )}
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
            {formatDate(currentVersion.updatedAt)}
          </div>
        </PropertyRow>
        {currentVersion.publishedAt && (
          <PropertyRow label={__("Published Date")}>
            <div className="text-sm text-txt-secondary">
              {formatDate(currentVersion.publishedAt)}
            </div>
          </PropertyRow>
        )}
      </Drawer>
    </>
  );
}

type Version = NodeOf<DocumentDetailPageDocumentFragment$data["versions"]>;

function EditablePropertyContent({
  children,
  onSave,
  onCancel,
  disabled,
}: {
  children: React.ReactNode;
  onSave: () => void;
  onCancel: () => void;
  disabled?: boolean;
}) {
  return (
    <div className="flex items-center gap-2">
      <div className="flex-1">{children}</div>
      <Button
        variant="quaternary"
        icon={IconCheckmark1}
        onClick={onSave}
        disabled={disabled}
      />
      <Button variant="quaternary" icon={IconCrossLargeX} onClick={onCancel} />
    </div>
  );
}

function ReadOnlyPropertyContent({
  children,
  onEdit,
  canEdit = true,
}: {
  children: React.ReactNode;
  onEdit: () => void;
  canEdit?: boolean;
}) {
  return (
    <div className="flex items-center justify-between gap-3">
      {children}
      {canEdit && (
        <Button variant="quaternary" icon={IconPencil} onClick={onEdit} />
      )}
    </div>
  );
}

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
  const { dateTimeFormat, __ } = useTranslate();
  const organizationId = useOrganizationId();
  const suffix = useLocation().pathname.split("/").at(-1);
  return (
    <Link
      to={`/organizations/${organizationId}/documents/${document.id}/versions/${version.id}/${suffix}`}
      onClick={() => onSelect(version)}
      className="flex items-center gap-2 py-2 px-[10px] w-full hover:bg-tertiary-hover cursor-pointer rounded"
      {...props}
    >
      <div className="flex gap-3 w-full overflow-hidden">
        <div
          className={clsx(
            "shrink-0 flex items-center justify-center size-10",
            active && "bg-active rounded",
          )}
        >
          <div className="text-base text-txt-primary whitespace-nowrap font-bold text-center">
            {version.version}
          </div>
        </div>
        <div className="flex-1 space-y-[2px] overflow-hidden">
          <div className="flex items-center gap-2 overflow-hidden">
            {version.status === "DRAFT" && (
              <Badge variant="neutral" size="sm">
                {__("Draft")}
              </Badge>
            )}
          </div>
          <div className="text-xs text-txt-secondary whitespace-nowrap overflow-hidden text-ellipsis">
            {dateTimeFormat(version.publishedAt ?? version.updatedAt)}
          </div>
        </div>
      </div>
    </Link>
  );
}
