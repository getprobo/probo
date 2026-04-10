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
import { Badge, Breadcrumb, Button, IconUpload, PageHeader, TabBadge, TabLink, Tabs } from "@probo/ui";
import { useCallback, useRef, useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet, useLocation, useNavigate, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { DocumentLayoutQuery } from "#/__generated__/core/DocumentLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { DocumentActionsDropdown } from "./_components/DocumentActionsDropdown";
import { DocumentDetailsCard } from "./_components/DocumentDetailsCard";
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
        ...DocumentDetailsCard_versionFragment
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
        writeMode
        canPublish: permission(action: "core:document-version:publish")
        ...PublishDialog_documentFragment
        controlInfo: controls(first: 0) {
          totalCount
        }
        ...DocumentActionsDropdown_documentFragment
        ...DocumentDetailsCard_documentFragment
        lastVersion: versions(first: 1 orderBy: { field: CREATED_AT, direction: DESC })
        @connection(key: "DocumentLayout_lastVersion") {
          edges {
            node {
              id
              title
              status
              ...DocumentTitleFormFragment
              ...DocumentActionsDropdown_versionFragment
              ...DocumentDetailsCard_versionFragment
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
  const navigate = useNavigate();
  const location = useLocation();

  const { __ } = useTranslate();

  const publishDialogRef = useRef<PublishDialogRef>(null);
  const [approvalRequestedAt, setApprovalRequestedAt] = useState(0);
  const [versionChangedAt, setVersionChangedAt] = useState(0);

  const handlePublishOrApproval = useCallback(() => {
    onRefetch();
    setApprovalRequestedAt(Date.now());
  }, [onRefetch]);

  const { document, version } = usePreloadedQuery<DocumentLayoutQuery>(documentLayoutQuery, queryRef);
  if (document.__typename !== "Document" || (version && version.__typename !== "DocumentVersion")) {
    throw new Error("invalid node type");
  }
  const lastVersion = document.lastVersion?.edges[0]?.node;

  if (!version && !lastVersion) {
    throw new Error("current version not specified");
  }

  const currentVersion = version ?? lastVersion;
  const isLatestVersion = currentVersion.id === lastVersion?.id;
  const isPendingApproval = currentVersion.status === "PENDING_APPROVAL";
  const isDraft = currentVersion.status === "DRAFT";
  const isPublished = currentVersion.status === "PUBLISHED";
  const isGenerated = document.writeMode === "GENERATED";
  const isEditable = isLatestVersion && !isPendingApproval;
  const lastQuorum = currentVersion.approvalQuorums?.edges?.[0]?.node ?? null;
  const hasApprovals = lastQuorum != null;

  const currentTab = location.pathname.split("/").at(-1);

  // For changes on the current version (type, classification, title, content).
  // Refreshes layout data but does NOT remount the editor.
  const handleDocumentUpdated = useCallback(() => {
    if (versionId) {
      void navigate(
        `/organizations/${organizationId}/documents/${document.id}/${currentTab}`,
        { replace: true },
      );
    } else {
      onRefetch();
    }
  }, [versionId, currentTab, navigate, organizationId, document.id, onRefetch]);

  // For structural version changes (delete draft, revert).
  // Refreshes layout data AND remounts the editor via versionChangedAt.
  const handleVersionChanged = useCallback(() => {
    if (versionId) {
      void navigate(
        `/organizations/${organizationId}/documents/${document.id}/${currentTab}`,
        { replace: true },
      );
    } else {
      onRefetch();
      setVersionChangedAt(Date.now());
    }
  }, [versionId, currentTab, navigate, organizationId, document.id, onRefetch]);

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
            <DocumentVersionsDropdown currentTab={currentTab} />
            <DocumentActionsDropdown
              documentFragmentRef={document}
              versionFragmentRef={currentVersion}
              onVersionChanged={handleVersionChanged}
            />
          </div>
        </div>

        <PageHeader
          title={(
            <DocumentTitleForm
              fKey={currentVersion}
              documentId={document.id}
              documentStatus={document.status}
              isEditable={isEditable && !isGenerated}
              onDocumentUpdated={handleDocumentUpdated}
            />
          )}
        >
          {isGenerated && <Badge variant="neutral">{__("Generated")}</Badge>}
          <Badge
            variant={currentVersion.status === "PUBLISHED" ? "success" : currentVersion.status === "PENDING_APPROVAL" ? "warning" : "highlight"}
          >
            {currentVersion.status === "PUBLISHED" ? __("Published") : currentVersion.status === "PENDING_APPROVAL" ? __("Pending approval") : __("Draft")}
          </Badge>
        </PageHeader>

        <DocumentDetailsCard
          documentFragmentRef={document}
          versionFragmentRef={currentVersion}
          isEditable={isEditable}
          isGenerated={isGenerated}
          isLatestVersion={isLatestVersion}
          onDocumentUpdated={handleDocumentUpdated}
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
                {currentVersion.signedSignatures?.totalCount ?? 0}
                /
                {currentVersion.signatures?.totalCount ?? 0}
              </TabBadge>
            </TabLink>
          )}
        </Tabs>

        <Outlet
          context={{
            onRefetch,
            onDocumentUpdated: handleDocumentUpdated,
            approvalRequestedAt,
            versionChangedAt,
            isEditable,
          }}
        />
      </div>

      <PublishDialog
        ref={publishDialogRef}
        documentId={document.id}
        documentFragmentRef={document}
        onSuccess={handlePublishOrApproval}
      />
    </>
  );
}
