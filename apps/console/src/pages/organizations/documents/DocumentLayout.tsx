import { useTranslate } from "@probo/i18n";
import { Breadcrumb, Button, IconCheckmark1, PageHeader, TabBadge, TabLink, Tabs } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { DocumentLayoutQuery } from "#/__generated__/core/DocumentLayoutQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { DocumentActionsDropdownn } from "./_components/DocumentActionsDropdown";
import { DocumentLayoutDrawer } from "./_components/DocumentLayoutDrawer";
import { DocumentTitleForm } from "./_components/DocumentTitleForm";
import { DocumentVersionsDropdown } from "./_components/DocumentVersionsDropdown";

export const documentLayoutQuery = graphql`
  query DocumentLayoutQuery($documentId: ID! $versionId: ID! $versionSpecified: Boolean!) {
    # We use this on /documents/:documentId/versions/:versionId
    version: node(id: $versionId) @include(if: $versionSpecified) {
      __typename
      ... on DocumentVersion {
        id
        status
        ...DocumentActionsDropdown_versionFragment
        ...DocumentLayoutDrawer_versionFragment
        signatures(first: 0 filter: { activeContract: true }) {
          totalCount
        }
        signedSignatures: signatures(first: 0 filter: { states: [SIGNED], activeContract: true }) {
          totalCount
        }
      }
    }
    document: node(id: $documentId) {
      __typename
      ... on Document {
        id
        title
        canPublish: permission(action: "core:document-version:publish")
        controlInfo: controls(first: 0) {
          totalCount
        }
        ...DocumentTitleFormFragment
        ...DocumentActionsDropdown_documentFragment
        ...DocumentLayoutDrawer_documentFragment
        # We use this on /documents/:documentId
        lastVersion: versions(first: 1 orderBy: { field: CREATED_AT, direction: DESC }) @skip(if: $versionSpecified) {
          edges {
            node {
              id
              status
              ...DocumentActionsDropdown_versionFragment
              ...DocumentLayoutDrawer_versionFragment
              signatures(first: 0 filter: { activeContract: true }) {
                totalCount
              }
              signedSignatures: signatures(first: 0 filter: { states: [SIGNED], activeContract: true }) {
                totalCount
              }
            }
          }
        }
      }
    }
  }
`;

const publishDocumentVersionMutation = graphql`
  mutation DocumentLayout_publishVersionMutation(
    $input: PublishDocumentVersionInput!
  ) {
    publishDocumentVersion(input: $input) {
      document {
        id
      }
    }
  }
`;

export function DocumentLayout(props: { queryRef: PreloadedQuery<DocumentLayoutQuery> }) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const { versionId } = useParams();

  const { __ } = useTranslate();

  const { document, version } = usePreloadedQuery<DocumentLayoutQuery>(documentLayoutQuery, queryRef);
  if (document.__typename !== "Document" || (version && version.__typename !== "DocumentVersion")) {
    throw new Error("invalid node type");
  }
  const lastVersion = document.lastVersion?.edges[0].node;

  if (!version && !lastVersion) {
    throw new Error("current version not specified");
  }

  // It is ok to cas as NonNullable here since we know we have either version or lastVersion
  const currentVersion = version ?? lastVersion as NonNullable<typeof version | typeof lastVersion>;
  const isDraft = currentVersion.status === "DRAFT";

  const [publishDocumentVersion, isPublishing] = useMutationWithToasts(
    publishDocumentVersionMutation,
    {
      successMessage: __("Document published successfully."),
      errorMessage: __("Failed to publish document"),
    },
  );

  const handlePublish = async () => {
    await publishDocumentVersion({
      variables: {
        input: { documentId: document.id },
      },
    });
  };

  const urlPrefix = versionId
    ? `/organizations/${organizationId}/documents/${document.id}/versions/${versionId}`
    : `/organizations/${organizationId}/documents/${document.id}`;

  return (
    <>
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
            <DocumentVersionsDropdown />
            <DocumentActionsDropdownn documentFragmentRef={document} versionFragmentRef={currentVersion} />
          </div>
        </div>

        <PageHeader
          title={<DocumentTitleForm fKey={document} />}
        />

        <Tabs>
          <TabLink to={`${urlPrefix}/description`}>{__("Description")}</TabLink>
          <TabLink to={`${urlPrefix}/controls`}>
            {__("Controls")}
            <TabBadge>{document.controlInfo.totalCount}</TabBadge>
          </TabLink>
          {!isDraft && (
            <TabLink to={`${urlPrefix}/signatures`}>
              {__("Signatures")}
              <TabBadge>
                {currentVersion.signedSignatures.totalCount}
                /
                {currentVersion.signatures.totalCount}
              </TabBadge>
            </TabLink>
          )}
        </Tabs>

        <Outlet />
      </div>

      <DocumentLayoutDrawer documentFragmentRef={document} versionFragmentRef={currentVersion} />
    </>
  );
}
