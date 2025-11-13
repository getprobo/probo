import {
  graphql,
  useFragment,
  useLazyLoadQuery,
} from "react-relay";
import type { DocumentGraphNodeQuery } from "/hooks/graph/__generated__/DocumentGraphNodeQuery.graphql";
import {
  documentNodeQuery,
  useDeleteDocumentMutation,
  useDeleteDraftDocumentVersionMutation,
} from "/hooks/graph/DocumentGraph";
import { usePageTitle } from "@probo/hooks";
import type {
  DocumentDetailPageDocumentFragment$data,
  DocumentDetailPageDocumentFragment$key,
} from "./__generated__/DocumentDetailPageDocumentFragment.graphql";
import type { DocumentDetailPageExportPDFMutation } from "./__generated__/DocumentDetailPageExportPDFMutation.graphql";
import type { DocumentDetailPageUpdateMutation } from "./__generated__/DocumentDetailPageUpdateMutation.graphql";
import type { DocumentDetailPageUserEmailQuery } from "./__generated__/DocumentDetailPageUserEmailQuery.graphql";
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
  IconPencil,
  IconTrashCan,
  IconCrossLargeX,
  Input,
  PageHeader,
  PropertyRow,
  TabBadge,
  TabLink,
  Tabs,
  useConfirm,
  Spinner,
} from "@probo/ui";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import {
  getDocumentTypeLabel,
  getDocumentClassificationLabel,
  sprintf,
  documentTypes,
  documentClassifications,
  formatDate,
} from "@probo/helpers";
import {
  Link,
  Outlet,
  useLocation,
  useNavigate,
  useParams,
} from "react-router";
import UpdateVersionDialog from "./dialogs/UpdateVersionDialog";
import {
  PdfDownloadDialog,
  type PdfDownloadDialogRef,
} from "/components/documents/PdfDownloadDialog";
import { useRef, useState, useEffect } from "react";
import type { NodeOf } from "/types.ts";
import clsx from "clsx";
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import { ControlledField } from "/components/form/ControlledField";
import { DocumentTypeOptions } from "/components/form/DocumentTypeOptions";
import { DocumentClassificationOptions } from "/components/form/DocumentClassificationOptions";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { Authorized, isAuthorized, fetchPermissions } from "/permissions";

const documentFragment = graphql`
  fragment DocumentDetailPageDocumentFragment on Document
  @argumentDefinitions(
    includeControls: { type: "Boolean!", defaultValue: false }
    includeSignatures: { type: "Boolean!", defaultValue: false }
    useRequestedVersions: { type: "Boolean!", defaultValue: false }
  ) {
    id
    title
    documentType
    classification
    organization {
      id
    }
    owner {
      id
      fullName
    }
    controlsInfo: controls(first: 0) @include(if: $includeControls) {
      totalCount
    }
    versions(first: 20) @skip(if: $useRequestedVersions) @connection(key: "DocumentDetailPage_versions") {
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
          signatures(first: 1000) @include(if: $includeSignatures)
            @connection(key: "DocumentDetailPage_signatures", filters: []) {
            __id
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
    requestedVersions(first: 20) @include(if: $useRequestedVersions) @connection(key: "DocumentDetailPage_requestedVersions") {
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
          signatures(first: 1000) @include(if: $includeSignatures)
            @connection(key: "DocumentDetailPage_signatures", filters: []) {
            __id
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

graphql`
  fragment DocumentDetailPageRowFragment on Document {
    id
    title
    description
    documentType
    classification
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

const publishDocumentVersionMutation = graphql`
  mutation DocumentDetailPagePublishMutation(
    $input: PublishDocumentVersionInput!
  ) {
    publishDocumentVersion(input: $input) {
      document {
        id
        ...DocumentDetailPageRowFragment
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

const UserEmailQuery = graphql`
  query DocumentDetailPageUserEmailQuery {
    viewer {
      user {
        email
      }
    }
  }
`;

function DocumentDetailPageContent() {
  const { versionId } = useParams<{ versionId?: string }>();
  const { documentId } = useParams<{ documentId: string }>();
  const organizationId = useOrganizationId();
  const canViewControls = isAuthorized(organizationId, "Document", "listControls");
  const canViewSignatures = isAuthorized(organizationId, "DocumentVersion", "signatures");
  const canListVersions = isAuthorized(organizationId, "Document", "listVersions");

  const { node } = useLazyLoadQuery<DocumentGraphNodeQuery>(
    documentNodeQuery,
    {
      documentId: documentId!,
      includeControls: canViewControls,
      includeSignatures: canViewSignatures,
      useRequestedVersions: !canListVersions
    },
    { fetchPolicy: 'store-or-network' }
  );

  const document = useFragment<DocumentDetailPageDocumentFragment$key>(
    documentFragment,
    node
  );
  const { __ } = useTranslate();
  const navigate = useNavigate();

  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [isEditingOwner, setIsEditingOwner] = useState(false);
  const [isEditingType, setIsEditingType] = useState(false);
  const [isEditingClassification, setIsEditingClassification] = useState(false);
  const versionsConnection = canListVersions ? document.versions : document.requestedVersions;
  const versions = versionsConnection?.edges.map((edge: any) => edge.node) || [];
  const currentVersion =
    versionsConnection?.edges.find((v: any) => v.node.id === versionId)?.node ??
    versionsConnection?.edges[0]?.node;

  if (!currentVersion) {
    return null;
  }

  const isDraft = currentVersion.status === "DRAFT";

  const controlsCount = document.controlsInfo?.totalCount ?? 0;
  const signatures = currentVersion.signatures?.edges?.map((s) => s.node) ?? [];
  const signedCount = signatures.filter((s) => s.state === "SIGNED").length;
  const [publishDocumentVersion, isPublishing] = useMutationWithToasts(
    publishDocumentVersionMutation,
    {
      successMessage: __("Document published successfully."),
      errorMessage: __("Failed to publish document"),
    }
  );
  const [deleteDocument, isDeleting] = useDeleteDocumentMutation();
  const [deleteDraftDocumentVersion, isDeletingDraft] =
    useDeleteDraftDocumentVersionMutation();
  const [exportDocumentVersionPDF, isExporting] =
    useMutationWithToasts<DocumentDetailPageExportPDFMutation>(
      exportDocumentVersionPDFMutation,
      {
        successMessage: __("PDF download started."),
        errorMessage: __("Failed to generate PDF"),
      }
    );

  const userEmailData = useLazyLoadQuery<DocumentDetailPageUserEmailQuery>(UserEmailQuery, {});
  const defaultEmail = userEmailData.viewer.user.email;
  const [updateDocument, isUpdatingDocument] = useMutationWithToasts<DocumentDetailPageUpdateMutation>(
    updateDocumentMutation,
    {
      successMessage: __("Document updated successfully."),
      errorMessage: __("Failed to update document"),
    }
  );
  const versionConnectionId = versionsConnection?.__id || "";

  const { register, control, handleSubmit, reset } = useFormWithSchema(
    documentUpdateSchema,
    {
      defaultValues: {
        title: document.title,
        ownerId: currentVersion.owner?.id || "",
        documentType: document.documentType,
        classification: currentVersion.classification,
      },
    }
  );

  usePageTitle(document.title);

  const handleUpdateTitle = (data: { title: string }) => {
    updateDocument({
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

  const handleUpdateOwner = (data: { ownerId: string }) => {
    updateDocument({
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

  const handleUpdateDocumentType = (data: {
    documentType: (typeof documentTypes)[number];
  }) => {
    updateDocument({
      variables: {
        input: {
          id: document.id,
          documentType: data.documentType,
        },
      },
      onSuccess: () => {
        setIsEditingType(false);
      },
    });
  };

  const handleUpdateClassification = (data: {
    classification: (typeof documentClassifications)[number];
  }) => {
    updateDocument({
      variables: {
        input: {
          id: document.id,
          classification: data.classification,
        },
      },
      onSuccess: () => {
        setIsEditingClassification(false);
      },
    });
  };

  const handlePublish = () => {
    publishDocumentVersion({
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
          deleteDocument({
            variables: {
              input: { documentId: document.id },
            },
            onSuccess() {
              navigate(`/organizations/${organizationId}/documents`);
              resolve();
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

  const handleDeleteDraft = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteDraftDocumentVersion({
            variables: {
              input: { documentVersionId: currentVersion.id },
              connections: [versionConnectionId],
            },
            onSuccess() {
              resolve();
            },
            onError: () => resolve(),
          });
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete the draft version %s of "%s". This action cannot be undone.'
          ),
          currentVersion.version,
          document.title
        ),
      }
    );
  };

  const handleDownloadPdf = (options: {
    withWatermark: boolean;
    withSignatures: boolean;
    watermarkEmail?: string;
  }) => {
    const input = {
      documentVersionId: currentVersion.id,
      withWatermark: options.withWatermark,
      withSignatures: options.withSignatures,
      ...(options.withWatermark &&
        options.watermarkEmail && { watermarkEmail: options.watermarkEmail }),
    };

    exportDocumentVersionPDF({
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
  const urlPrefix = versionId
    ? `/organizations/${organizationId}/documents/${document.id}/versions/${versionId}`
    : `/organizations/${organizationId}/documents/${document.id}`;

  const handleDownloadClick = () => {
    if (!canViewSignatures) {
      handleDownloadPdf({
        withWatermark: true,
        withSignatures: false,
        watermarkEmail: defaultEmail,
      });
    } else {
      pdfDownloadDialogRef.current?.open();
    }
  };

  return (
    <>
      <UpdateVersionDialog
        ref={updateDialogRef}
        document={document}
        connectionId={versionConnectionId}
      />
      <PdfDownloadDialog
        ref={pdfDownloadDialogRef}
        onDownload={handleDownloadPdf}
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
            {isDraft && (
              <Authorized entity="Document" action="publishDocumentVersion">
                <Button
                  onClick={handlePublish}
                  icon={IconCheckmark1}
                  disabled={isPublishing}
                >
                  {__("Publish")}
                </Button>
              </Authorized>
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
            </Dropdown>

            <ActionDropdown variant="secondary">
              <Authorized entity="Document" action="updateDocument">
                <DropdownItem
                  onClick={() => updateDialogRef.current?.open()}
                  icon={IconPencil}
                >
                  {isDraft ? __("Edit draft document") : __("Create new draft")}
                </DropdownItem>
              </Authorized>
              {isDraft && versions.length > 1 && (
                <Authorized entity="Document" action="deleteDocument">
                  <DropdownItem
                    onClick={handleDeleteDraft}
                    icon={IconTrashCan}
                    disabled={isDeletingDraft}
                  >
                    {__("Delete draft document")}
                  </DropdownItem>
                </Authorized>
              )}
              <DropdownItem
                onClick={handleDownloadClick}
                icon={IconArrowDown}
                disabled={isExporting}
              >
                {__("Download PDF")}
              </DropdownItem>
              <Authorized entity="Document" action="deleteDocument">
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  disabled={isDeleting}
                  onClick={handleDelete}
                >
                  {__("Delete document")}
                </DropdownItem>
              </Authorized>
            </ActionDropdown>
          </div>
        </div>
        <PageHeader
          title={
            isEditingTitle ? (
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
                      handleSubmit(handleUpdateTitle)();
                    }
                  }}
                />
                <Button
                  variant="quaternary"
                  icon={IconCheckmark1}
                  onClick={handleSubmit(handleUpdateTitle)}
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
            ) : (
              <div className="flex items-center gap-2">
                <span>{document.title}</span>
                <Authorized entity="Document" action="updateDocument">
                  <Button
                    variant="quaternary"
                    icon={IconPencil}
                    onClick={() => setIsEditingTitle(true)}
                  />
                </Authorized>
              </div>
            )
          }
        />

        <Tabs>
          <TabLink to={`${urlPrefix}/description`}>{__("Content")}</TabLink>
          <Authorized entity="Document" action="listControls">
            <TabLink to={`${urlPrefix}/controls`}>
              {__("Controls")}
              {document.controlsInfo && <TabBadge>{controlsCount}</TabBadge>}
            </TabLink>
          </Authorized>
          {!isDraft && (
            <Authorized entity="DocumentVersion" action="signatures">
              <TabLink to={`${urlPrefix}/signatures`}>
                {__("Signatures")}
                {currentVersion.signatures && (
                  <TabBadge>
                    {signedCount}/{signatures.length}
                  </TabBadge>
                )}
              </TabLink>
            </Authorized>
          )}
        </Tabs>

        <Outlet context={{ document, version: currentVersion }} />
      </div>
      <Drawer>
        <div className="text-base text-txt-primary font-medium mb-4">
          {__("Properties")}
        </div>
        <PropertyRow label={__("Owner")}>
          {isEditingOwner ? (
            <EditablePropertyContent
              onSave={handleSubmit(handleUpdateOwner)}
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
          ) : (
            <ReadOnlyPropertyContent onEdit={() => setIsEditingOwner(true)}>
              <Badge variant="highlight" size="md" className="gap-2">
                <Avatar name={currentVersion.owner?.fullName ?? ""} />
                {currentVersion.owner?.fullName}
              </Badge>
            </ReadOnlyPropertyContent>
          )}
        </PropertyRow>
        <PropertyRow label={__("Type")}>
          {isEditingType ? (
            <EditablePropertyContent
              onSave={handleSubmit(handleUpdateDocumentType)}
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
          ) : (
            <ReadOnlyPropertyContent onEdit={() => setIsEditingType(true)}>
              <div className="text-sm text-txt-secondary">
                {getDocumentTypeLabel(__, document.documentType)}
              </div>
            </ReadOnlyPropertyContent>
          )}
        </PropertyRow>
        <PropertyRow label={__("Classification")}>
          {isEditingClassification ? (
            <EditablePropertyContent
              onSave={handleSubmit(handleUpdateClassification)}
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
          ) : (
            <ReadOnlyPropertyContent
              onEdit={() => setIsEditingClassification(true)}
            >
              <div className="text-sm text-txt-secondary">
                {getDocumentClassificationLabel(__, currentVersion.classification)}
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
}: {
  children: React.ReactNode;
  onEdit: () => void;
}) {
  return (
    <div className="flex items-center justify-between gap-3">
      {children}
      <Authorized entity="Document" action="updateDocument">
        <Button variant="quaternary" icon={IconPencil} onClick={onEdit} />
      </Authorized>
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
            "flex-shrink-0 flex items-center justify-center size-10",
            active && "bg-active rounded"
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

export default function DocumentDetailPage() {
  const organizationId = useOrganizationId();

  const [cacheKey, setCacheKey] = useState(0);

  useEffect(() => {
    fetchPermissions(organizationId).then(() => {
      setCacheKey(prev => prev + 1);
    });
  }, [organizationId]);

  const hasCheckedPermissions = cacheKey > 0;

  if (!hasCheckedPermissions) {
    return <Spinner />;
  }

  return <DocumentDetailPageContent key={cacheKey} />;
}
