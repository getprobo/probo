// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { usePageTitle } from "@probo/hooks";
import { Badge, Button, IconUpload, PageHeader, TabBadge, TabLink, Tabs } from "@probo/ui";
import { useCallback, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
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
        status
        title
        ...DocumentActionsDropdown_versionFragment
        ...DocumentDetailsCard_versionFragment
        signatures(first: 0 filter: { activeContract: true, profileStates: [ACTIVE] }) {
          totalCount
        }
        signedSignatures: signatures(first: 0 filter: { states: [SIGNED], activeContract: true, profileStates: [ACTIVE] }) {
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
        title
        status
        writeMode
        canUpdate: permission(action: "core:document:update")
        canPublish: permission(action: "core:document-version:publish")
        ...PublishDialog_documentFragment
        controlInfo: controls(first: 0) {
          totalCount
        }
        ...DocumentActionsDropdown_documentFragment
        ...DocumentDetailsCard_documentFragment
        pendingApproval: versions(first: 1, filter: { statuses: [PENDING_APPROVAL] }) {
          edges {
            node {
              id
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
        publishedVersion: versions(
          first: 1
          filter: { statuses: [PUBLISHED] }
          orderBy: { field: CREATED_AT, direction: DESC }
        ) {
          edges {
            node {
              id
              ...DocumentActionsDropdown_versionFragment
              signatures(first: 0 filter: { activeContract: true, profileStates: [ACTIVE] }) {
                totalCount
              }
              signedSignatures: signatures(first: 0 filter: { states: [SIGNED], activeContract: true, profileStates: [ACTIVE] }) {
                totalCount
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

  const { t } = useTranslation();

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
  const snapshot = version ?? null;
  const pending = document.pendingApproval.edges[0]?.node ?? null;
  const published = document.publishedVersion.edges[0]?.node ?? null;
  const isLive = snapshot == null;
  const isGenerated = document.writeMode === "GENERATED";
  const isEditable = isLive
    && pending == null
    && document.status !== "ARCHIVED"
    && !isGenerated;
  const title = snapshot?.title ?? document.title;
  const approvalSource = snapshot ?? pending;
  const lastQuorum = approvalSource?.approvalQuorums?.edges?.[0]?.node ?? null;
  const hasApprovals = lastQuorum != null;
  const signatureSource = snapshot?.status === "PUBLISHED" ? snapshot : isLive ? published : null;

  const currentTab = location.pathname.split("/").at(-1);

  usePageTitle(title);

  // For changes on the current version (type, classification, title, content).
  // Refreshes layout data but does NOT remount the editor.
  const handleDocumentUpdated = useCallback(() => {
    if (versionId) {
      void navigate(
        `/organizations/${organizationId}/governance/documents/${document.id}/${currentTab}`,
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
        `/organizations/${organizationId}/governance/documents/${document.id}/${currentTab}`,
        { replace: true },
      );
    } else {
      onRefetch();
      setVersionChangedAt(Date.now());
    }
  }, [versionId, currentTab, navigate, organizationId, document.id, onRefetch]);

  const urlPrefix = versionId
    ? `/organizations/${organizationId}/governance/documents/${document.id}/versions/${versionId}`
    : `/organizations/${organizationId}/governance/documents/${document.id}`;

  return (
    <>
      <div className="flex flex-col gap-6 h-full">
        <div className="flex justify-end items-center mb-4">
          <div className="flex gap-2">
            {isLive && document.canPublish && pending == null && (
              <Button
                icon={IconUpload}
                onClick={() => publishDialogRef.current?.open()}
              >
                {t("documentLayout.actions.publish")}
              </Button>
            )}
            <DocumentVersionsDropdown currentTab={currentTab} />
            <DocumentActionsDropdown
              documentFragmentRef={document}
              versionFragmentRef={snapshot ?? published}
              isLive={isLive}
              onVersionChanged={handleVersionChanged}
            />
          </div>
        </div>

        <PageHeader
          title={(
            <DocumentTitleForm
              title={title}
              canUpdate={document.canUpdate}
              documentId={document.id}
              documentStatus={document.status}
              isEditable={isEditable}
              onDocumentUpdated={handleDocumentUpdated}
            />
          )}
        >
          {isGenerated && (
            <Badge variant="neutral">{t("documentLayout.status.generated")}</Badge>
          )}
          <Badge variant={snapshotStatusVariant(snapshot?.status)}>
            {snapshotStatusLabel(t, snapshot?.status)}
          </Badge>
        </PageHeader>

        <DocumentDetailsCard
          documentFragmentRef={document}
          versionFragmentRef={snapshot}
          isEditable={isEditable}
          isLatestVersion={isLive}
          hasPendingApproval={pending != null}
          onDocumentUpdated={handleDocumentUpdated}
        />

        <Tabs>
          <TabLink to={`${urlPrefix}/description`}>
            {t("documentLayout.tabs.description")}
          </TabLink>
          <TabLink to={`${urlPrefix}/controls`}>
            {t("documentLayout.tabs.controls")}
            <TabBadge>{document.controlInfo.totalCount}</TabBadge>
          </TabLink>
          {hasApprovals && (
            <TabLink to={`${urlPrefix}/approvals`}>
              {t("documentLayout.tabs.approvals")}
              <TabBadge>
                {lastQuorum?.status === "REJECTED"
                  ? t("documentLayout.status.rejected")
                  : `${lastQuorum?.approvedDecisions.totalCount ?? 0}/${lastQuorum?.decisions.totalCount ?? 0}`}
              </TabBadge>
            </TabLink>
          )}
          {signatureSource && (
            <TabLink to={`${urlPrefix}/signatures`}>
              {t("documentLayout.tabs.signatures")}
              <TabBadge>
                {signatureSource.signedSignatures?.totalCount ?? 0}
                /
                {signatureSource.signatures?.totalCount ?? 0}
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

function snapshotStatusVariant(status: string | undefined) {
  switch (status) {
    case "PUBLISHED":
      return "success" as const;
    case "PENDING_APPROVAL":
      return "warning" as const;
    case "REJECTED":
    case "VOIDED":
      return "danger" as const;
    default:
      return "highlight" as const;
  }
}

function snapshotStatusLabel(
  t: (key: string) => string,
  status: string | undefined,
) {
  switch (status) {
    case "PUBLISHED":
      return t("documentLayout.status.published");
    case "PENDING_APPROVAL":
      return t("documentLayout.status.pendingApproval");
    case "REJECTED":
      return t("documentLayout.status.rejected");
    case "VOIDED":
      return t("documentLayout.status.voided");
    default:
      return t("documentLayout.status.draft");
  }
}
