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
  query DocumentLayoutQuery($documentId: ID!) {
    document: node(id: $documentId) {
      __typename
      ... on Document {
        id
        title
        classification
        owner {
          id
          fullName
        }
        canPublish: permission(action: "core:document-version:publish")
        controlInfo: controls(first: 0) {
          totalCount
        }
        ...DocumentTitleFormFragment
        ...DocumentActionsDropdownFragment
        ...DocumentLayoutDrawerFragment
        ...DocumentControlsTabFragment
        versions(first: 20) @connection(key: "DocumentLayout_versions") {
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

  const { document } = usePreloadedQuery<DocumentLayoutQuery>(documentLayoutQuery, queryRef);
  if (document.__typename !== "Document") {
    throw new Error("invalid node type");
  }

  const currentVersion
    = document.versions.edges.find(v => v.node.id === versionId)?.node
      ?? document.versions.edges[0].node;
  const isDraft = currentVersion.status === "DRAFT";
  const signatures = currentVersion.signatures?.edges?.map(s => s.node) ?? [];
  const signedSignatures = signatures.filter(s => s.state === "SIGNED");

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
            <DocumentVersionsDropdown currentVersionId={currentVersion.id} />
            <DocumentActionsDropdownn isDraft={isDraft} fKey={document} currentVersion={currentVersion} />
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
                {signedSignatures.length}
                /
                {signatures.length}
              </TabBadge>
            </TabLink>
          )}
        </Tabs>

        <Outlet context={{ document, version: currentVersion }} />
      </div>

      <DocumentLayoutDrawer fKey={document} currentVersion={currentVersion} isDraft={isDraft} />
    </>
  );
}
