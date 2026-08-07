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

import { formatError } from "@probo/helpers";
import {
  createRichEditorAutomergeDocument,
  RichEditor,
  supportsRichEditorCollaboration,
  useToast,
} from "@probo/ui";
import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { useOutletContext } from "react-router";
import { graphql } from "relay-runtime";
import { useDebounceCallback } from "usehooks-ts";

import type { DocumentDescriptionPage_updateContentMutation } from "#/__generated__/core/DocumentDescriptionPage_updateContentMutation.graphql";
import type { DocumentDescriptionPageQuery } from "#/__generated__/core/DocumentDescriptionPageQuery.graphql";

import {
  type AutomergeDocumentHandle,
  connectAutomergeDocument,
} from "./_lib/AutomergeDocumentHandle";

const autoSaveIntervalMs = 1000;

type CollaborationState = {
  versionID: string;
  handle?: AutomergeDocumentHandle;
  failed?: boolean;
  established?: boolean;
};

export const documentDescriptionPageQuery = graphql`
  query DocumentDescriptionPageQuery($documentId: ID! $versionId: ID! $versionSpecified: Boolean!) {
    # We use this on /documents/:documentId/versions/:versionId/description
    version: node(id: $versionId) @include(if: $versionSpecified) {
      __typename
      ... on DocumentVersion {
        id
        content
        status
      }
    }
    document: node(id: $documentId) {
      __typename
      ... on Document {
        id
        status
        writeMode
        canUpdate: permission(action: "core:document:update")
        # We use this on /documents/:documentId/description
        lastVersion: versions(first: 1 orderBy: { field: CREATED_AT, direction: DESC }) @skip(if: $versionSpecified) {
          edges {
            node {
              id
              content
              status
            }
          }
        }
      }
    }
  }
`;

const updateContentMutation = graphql`
  mutation DocumentDescriptionPage_updateContentMutation($input: UpdateDocumentInput!) {
    updateDocument(input: $input) {
      document {
        id
      }
      documentVersion {
        id
        content
        status
      }
    }
  }
`;

export function DocumentDescriptionPage(props: {
  queryRef: PreloadedQuery<DocumentDescriptionPageQuery>;
  versionChangedAt: number;
}) {
  const { queryRef, versionChangedAt } = props;

  const { t } = useTranslation();
  const { toast } = useToast();
  const { onDocumentUpdated, isEditable } = useOutletContext<{
    onDocumentUpdated: () => void;
    isEditable: boolean;
  }>();

  const { document, version } = usePreloadedQuery<DocumentDescriptionPageQuery>(
    documentDescriptionPageQuery,
    queryRef,
  );
  if (document.__typename !== "Document" || (version && version.__typename !== "DocumentVersion")) {
    throw new Error("invalid type for node");
  }

  const lastVersion = document.lastVersion?.edges[0].node;
  const currentVersion = lastVersion ?? version;
  if (!currentVersion) {
    throw new Error("Document version not found");
  }

  const [updateContent] = useMutation<DocumentDescriptionPage_updateContentMutation>(updateContentMutation);

  const documentId = document.id;
  const wasDraft = currentVersion.status === "DRAFT";

  const handleUpdate = useDebounceCallback(
    useCallback((content: string) => {
      updateContent({
        variables: {
          input: {
            id: documentId,
            content,
          },
        },
        onCompleted: (data, errors) => {
          if (errors?.length) {
            toast({
              title: t("documentDescriptionPage.errors.title"),
              description: formatError(t("documentDescriptionPage.errors.save"), errors),
              variant: "error",
            });
            return;
          }

          const draftReturned = !!data.updateDocument.documentVersion;
          if (wasDraft !== draftReturned) {
            onDocumentUpdated();
          }

          toast({
            title: t("documentDescriptionPage.messages.successTitle"),
            description: t("documentDescriptionPage.messages.saved"),
            variant: "success",
          });
        },
        onError: (error) => {
          toast({
            title: t("documentDescriptionPage.errors.title"),
            description: error.message ?? t("documentDescriptionPage.errors.save"),
            variant: "error",
          });
        },
      });
    }, [documentId, wasDraft, updateContent, toast, t, onDocumentUpdated]),
    autoSaveIntervalMs,
  );

  const canEdit = isEditable
    && document.canUpdate
    && document.status !== "ARCHIVED"
    && document.writeMode !== "GENERATED";
  const collaborationSupported = canEdit
    && supportsRichEditorCollaboration(currentVersion.content);
  const [collaboration, setCollaboration] = useState<CollaborationState>();

  useEffect(() => {
    if (!collaborationSupported) return;

    let cancelled = false;
    let activeHandle: AutomergeDocumentHandle | undefined;

    void connectAutomergeDocument(
      currentVersion.id,
      content => createRichEditorAutomergeDocument(content),
      () => {
        if (cancelled) return;

        setCollaboration({
          versionID: currentVersion.id,
          failed: true,
          established: true,
        });
        toast({
          title: t("documentDescriptionPage.errors.title"),
          description: t("documentDescriptionPage.errors.save"),
          variant: "error",
        });
      },
    ).then((handle) => {
      if (cancelled) {
        handle.close();
        return;
      }

      activeHandle = handle;
      setCollaboration({
        versionID: currentVersion.id,
        handle,
        established: true,
      });
    }).catch((error) => {
      if (cancelled) return;

      setCollaboration({
        versionID: currentVersion.id,
        failed: true,
      });
      toast({
        title: t("documentDescriptionPage.errors.title"),
        description: error instanceof Error
          ? error.message
          : t("documentDescriptionPage.errors.save"),
        variant: "error",
      });
    });

    return () => {
      cancelled = true;
      activeHandle?.close();
    };
  }, [
    collaborationSupported,
    currentVersion.id,
    t,
    toast,
  ]);

  const collaborationConnecting = collaborationSupported
    && (
      collaboration?.versionID !== currentVersion.id
      || (!collaboration?.handle && !collaboration?.failed)
    );
  const collaborationHandle = collaborationSupported
    && collaboration?.versionID === currentVersion.id
    && !collaboration.failed
    ? collaboration.handle
    : undefined;
  const collaborationUnavailable = collaborationSupported
    && collaboration?.versionID === currentVersion.id
    && collaboration.failed
    && collaboration.established;

  // The editor key must change on explicit actions (delete draft, edit
  // title/type) but NOT on auto-save side effects (cursor preservation).
  // We track a "data generation" that only increments when an explicit
  // action (versionChangedAt change) is followed by fresh data arriving
  // (currentVersion.id change). This uses React's "adjust state during
  // render" pattern so we avoid refs-during-render and setState-in-effects.
  const [prevVCA, setPrevVCA] = useState(versionChangedAt);
  const [prevVersionId, setPrevVersionId] = useState(currentVersion.id);
  const [dataGeneration, setDataGeneration] = useState(0);
  const [pendingExplicit, setPendingExplicit] = useState(false);

  if (versionChangedAt !== prevVCA) {
    setPrevVCA(versionChangedAt);
    if (currentVersion.id !== prevVersionId) {
      // Both changed at once — data was already available.
      setPrevVersionId(currentVersion.id);
      setDataGeneration(g => g + 1);
      setPendingExplicit(false);
    } else {
      // Explicit action fired but data hasn't arrived yet.
      setPendingExplicit(true);
    }
  } else if (currentVersion.id !== prevVersionId) {
    setPrevVersionId(currentVersion.id);
    if (pendingExplicit) {
      // Fresh data arrived for a pending explicit action — remount.
      setDataGeneration(g => g + 1);
      setPendingExplicit(false);
    }
    // Otherwise auto-save changed the version — don't bump generation.
  }

  const editorKey = `${version?.id ?? document.id}-${dataGeneration}`;

  return (
    <RichEditor
      key={editorKey}
      className="flex-1"
      content={currentVersion.content}
      data-theme="document"
      disabled={!canEdit || collaborationConnecting || collaborationUnavailable}
      collaborationHandle={collaborationHandle}
      onChangeContent={handleUpdate}
    />
  );
}
