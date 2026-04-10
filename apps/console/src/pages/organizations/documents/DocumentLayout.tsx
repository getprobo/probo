// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import { Breadcrumb, Button, IconUpload, PageHeader, TabBadge, TabLink, Tabs } from "@probo/ui";
import { useCallback, useRef, useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { DocumentLayoutQuery } from "#/__generated__/core/DocumentLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { DocumentActionsDropdown } from "./_components/DocumentActionsDropdown";
import { DocumentLayoutDrawer } from "./_components/DocumentLayoutDrawer";
import { DocumentTitleForm } from "./_components/DocumentTitleForm";
import { DocumentVersionsDropdown } from "./_components/DocumentVersionsDropdown";
import { PublishDialog, type PublishDialogRef } from "./_components/PublishDialog";

export const documentLayoutQuery = graphql`
  query DocumentLayoutQuery($documentId: ID! $versionId: ID! $versionSpecified: Boolean!) {
    # We use this on /documents/:documentId/versions/:versionId
    version: node(id: $versionId) @include(if: $versionSpecified) {
      __typename
      ... on DocumentVersion {
        id
        title
        status
        ...DocumentTitleFormFragment
        ...DocumentActionsDropdown_versionFragment
        ...DocumentLayoutDrawer_versionFragment
        signatures(first: 0 filter: { activeContract: true }) {
          totalCount
        }
        signedSignatures: signatures(first: 0 filter: { states: [SIGNED], activeContract: true }) {
          totalCount
        }
        approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
          edges {
            node {
              id
              status
              decisions(first: 0) {
                totalCount
              }
              approvedDecisions: decisions(first: 0 filter: { states: [APPROVED] }) {
                totalCount
              }
            }
          }
        }
      }
    }
    document: node(id: $documentId) {
      __typename
      ... on Document {
        id
        status
        canPublish: permission(action: "core:document-version:publish")
        ...PublishDialog_documentFragment
        controlInfo: controls(first: 0) {
          totalCount
        }
        ...DocumentActionsDropdown_documentFragment
        ...DocumentLayoutDrawer_documentFragment
        # We use this on /documents/:documentId
        lastVersion: versions(first: 1 orderBy: { field: CREATED_AT, direction: DESC }) @skip(if: $versionSpecified) {
          edges {
            node {
              id
              title
              status
              ...DocumentTitleFormFragment
              ...DocumentActionsDropdown_versionFragment
              ...DocumentLayoutDrawer_versionFragment
              signatures(first: 0 filter: { activeContract: true }) {
                totalCount
              }
              signedSignatures: signatures(first: 0 filter: { states: [SIGNED], activeContract: true }) {
                totalCount
              }
              approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
                edges {
                  node {
                    id
                    status
                    decisions(first: 0) {
                      totalCount
                    }
                    approvedDecisions: decisions(first: 0 filter: { states: [APPROVED] }) {
                      totalCount
                    }
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

export function DocumentLayout(props: { queryRef: PreloadedQuery<DocumentLayoutQuery>; onRefetch: () => void }) {
  const { queryRef, onRefetch } = props;

  const organizationId = useOrganizationId();
  const { versionId } = useParams();

  const { __ } = useTranslate();

  const publishDialogRef = useRef<PublishDialogRef>(null);
  const [approvalRequestedAt, setApprovalRequestedAt] = useState(0);
  const [draftDeletedAt, setDraftDeletedAt] = useState(0);

  const handlePublishOrApproval = useCallback(() => {
    onRefetch();
    setApprovalRequestedAt(Date.now());
  }, [onRefetch]);

  const handleDraftDeleted = useCallback(() => {
    onRefetch();
    setDraftDeletedAt(Date.now());
  }, [onRefetch]);

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
  const isPublished = currentVersion.status === "PUBLISHED";
  const lastQuorum = currentVersion.approvalQuorums?.edges?.[0]?.node ?? null;
  const hasApprovals = lastQuorum != null;

  const urlPrefix = versionId
    ? `/organizations/${organizationId}/documents/${document.id}/versions/${versionId}`
    : `/organizations/${organizationId}/documents/${document.id}`;

  return (
    <>
      <div className="flex flex-col gap-6 h-full">
        <div className="flex justify-between items-center mb-4">
          <Breadcrumb
            items={[
              {
                label: __("Documents"),
                to: `/organizations/${organizationId}/documents`,
              },
              {
                label: currentVersion.title,
              },
            ]}
          />

          <div className="flex gap-2">
            {isDraft && document.canPublish && (
              <Button
                icon={IconUpload}
                onClick={() => publishDialogRef.current?.open()}
              >
                {__("Publish")}
              </Button>
            )}
            <DocumentVersionsDropdown />
            <DocumentActionsDropdown
              documentFragmentRef={document}
              versionFragmentRef={currentVersion}
              onDraftDeleted={handleDraftDeleted}
            />
          </div>
        </div>

        <PageHeader
          title={<DocumentTitleForm fKey={currentVersion} documentId={document.id} />}
        />

        <Tabs>
          <TabLink to={`${urlPrefix}/description`}>{__("Description")}</TabLink>
          <TabLink to={`${urlPrefix}/controls`}>
            {__("Controls")}
            <TabBadge>{document.controlInfo.totalCount}</TabBadge>
          </TabLink>
          {hasApprovals && (
            <TabLink to={`${urlPrefix}/approvals`}>
              {__("Approvals")}
              <TabBadge>
                {lastQuorum?.status === "REJECTED"
                  ? __("Rejected")
                  : `${lastQuorum?.approvedDecisions.totalCount ?? 0}/${lastQuorum?.decisions.totalCount ?? 0}`}
              </TabBadge>
            </TabLink>
          )}
          {isPublished && (
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

        <Outlet context={{ onRefetch, approvalRequestedAt, draftDeletedAt }} />
      </div>

      <DocumentLayoutDrawer documentFragmentRef={document} versionFragmentRef={currentVersion} onRefetch={onRefetch} />

      <PublishDialog
        ref={publishDialogRef}
        documentId={document.id}
        documentFragmentRef={document}
        onSuccess={handlePublishOrApproval}
      />
    </>
  );
}
