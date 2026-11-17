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

import { XIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { CardSkeleton } from "@probo/ui/src/v2/Card/CardSkeleton";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useLazyLoadQuery, usePaginationFragment } from "react-relay";

import type { DocumentControlsSection_document$key } from "#/__generated__/core/DocumentControlsSection_document.graphql";
import type { DocumentControlsSectionList_document$key } from "#/__generated__/core/DocumentControlsSectionList_document.graphql";
import type { DocumentControlsSectionListQuery } from "#/__generated__/core/DocumentControlsSectionListQuery.graphql";
import type { DocumentControlsSectionListRefetchQuery } from "#/__generated__/core/DocumentControlsSectionListRefetchQuery.graphql";
import type { DocumentControlsSection_attachMutation } from "#/__generated__/core/DocumentControlsSection_attachMutation.graphql";
import type { DocumentControlsSection_detachMutation } from "#/__generated__/core/DocumentControlsSection_detachMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { documentControls } from "../variants";

import { DocumentControlCombobox } from "./DocumentControlCombobox";

const documentControlsPageSize = 50;

const documentControlsSectionFragment = graphql`
  fragment DocumentControlsSection_document on Document {
    id
    organization {
      canListControls: permission(action: "core:control:list")
      canLinkControl: permission(action: "core:control:create-document-mapping")
    }
  }
`;

const documentControlsListQuery = graphql`
  query DocumentControlsSectionListQuery($documentId: ID!) {
    node(id: $documentId) {
      __typename
      ... on Document {
        ...DocumentControlsSectionList_document
      }
    }
  }
`;

const documentControlsListFragment = graphql`
  fragment DocumentControlsSectionList_document on Document
  @refetchable(queryName: "DocumentControlsSectionListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey" }
  ) {
    id
    controls(first: $first, after: $after)
      @connection(key: "DocumentControlsSection_controls") {
      __id
      edges {
        node {
          id
          name
          sectionTitle
          canUnlink: permission(action: "core:control:delete-document-mapping")
          framework {
            id
          }
        }
      }
    }
  }
`;

const attachControlMutation = graphql`
  mutation DocumentControlsSection_attachMutation(
    $input: CreateControlDocumentMappingInput!
    $connections: [ID!]!
  ) {
    createControlDocumentMapping(input: $input) {
      controlEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          sectionTitle
          canUnlink: permission(action: "core:control:delete-document-mapping")
          framework {
            id
          }
        }
      }
    }
  }
`;

const detachControlMutation = graphql`
  mutation DocumentControlsSection_detachMutation(
    $input: DeleteControlDocumentMappingInput!
    $connections: [ID!]!
  ) {
    deleteControlDocumentMapping(input: $input) {
      deletedControlId @deleteEdge(connections: $connections)
    }
  }
`;

interface DocumentControlsSectionProps {
  documentKey: DocumentControlsSection_document$key;
}

export function DocumentControlsSection({ documentKey }: DocumentControlsSectionProps) {
  const document = useFragment(documentControlsSectionFragment, documentKey);
  if (!document.organization.canListControls) {
    return null;
  }

  return (
    <Suspense fallback={<CardSkeleton size={4} />}>
      <DocumentControlsLoaded
        documentId={document.id}
        canLink={document.organization.canLinkControl}
      />
    </Suspense>
  );
}

function DocumentControlsLoaded({
  documentId,
  canLink,
}: {
  documentId: string;
  canLink: boolean;
}) {
  const data = useLazyLoadQuery<DocumentControlsSectionListQuery>(
    documentControlsListQuery,
    { documentId },
  );
  const documentKey = data.node?.__typename === "Document" ? data.node : null;
  if (documentKey == null) {
    return null;
  }

  return <DocumentControlsList documentKey={documentKey} canLink={canLink} />;
}

function DocumentControlsList({
  documentKey,
  canLink,
}: {
  documentKey: DocumentControlsSectionList_document$key;
  canLink: boolean;
}) {
  const { t } = useTranslation("organizations/documents");
  const organizationId = useOrganizationId();
  const {
    data: document,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    DocumentControlsSectionListRefetchQuery,
    DocumentControlsSectionList_document$key
  >(documentControlsListFragment, documentKey);
  const [attachControl, isAttaching] = useMutation<DocumentControlsSection_attachMutation>(
    attachControlMutation,
    { errorToast: t("details.errors.update") },
  );
  const [detachControl, isDetaching] = useMutation<DocumentControlsSection_detachMutation>(
    detachControlMutation,
    { errorToast: t("details.errors.update") },
  );
  const isUpdating = isAttaching || isDetaching;
  const linksIncomplete = hasNext || isLoadingNext;
  const { root, names, name, control } = documentControls();
  const controls = document.controls.edges.map(edge => edge.node);
  const linkedIds = new Set(controls.map(item => item.id));

  return (
    <Card variant="soft" size={2}>
      <div className={root()}>
        <Heading level={2} size={2} weight="medium">
          {t("details.controls.title")}
        </Heading>
        <div className={names()}>
          {controls.length > 0
            ? controls.map(item => (
                <div key={item.id} className={name()}>
                  <Link
                    size={1}
                    title={`${item.sectionTitle} ${item.name}`}
                    className={control()}
                    to={`/organizations/${organizationId}/governance/frameworks/${item.framework.id}/controls/${item.id}`}
                  >
                    {`${item.sectionTitle} ${item.name}`}
                  </Link>
                  {item.canUnlink && (
                    <IconButton
                      size={1}
                      variant="ghost"
                      color="neutral"
                      disabled={isUpdating || linksIncomplete}
                      className="shrink-0"
                      aria-label={t("details.controls.remove", {
                        name: `${item.sectionTitle} ${item.name}`,
                      })}
                      onClick={() => {
                        void detachControl({
                          variables: {
                            input: {
                              controlId: item.id,
                              documentId: document.id,
                            },
                            connections: [document.controls.__id],
                          },
                        }).catch(() => {
                          // Error toast is already shown by useMutation.
                        });
                      }}
                    >
                      <XIcon />
                    </IconButton>
                  )}
                </div>
              ))
            : (
                <Text size={1} color="faint">{t("details.controls.none")}</Text>
              )}
          {hasNext && (
            <Button
              variant="ghost"
              color="neutral"
              size={1}
              className="self-start"
              loading={isLoadingNext}
              onClick={() => loadNext(documentControlsPageSize)}
            >
              {t("details.controls.loadMore")}
            </Button>
          )}
          {canLink && (
            <DocumentControlCombobox
              excludeIds={linkedIds}
              disabled={isUpdating || linksIncomplete}
              onSelect={(item) => {
                void attachControl({
                  variables: {
                    input: {
                      controlId: item.id,
                      documentId: document.id,
                    },
                    connections: [document.controls.__id],
                  },
                }).catch(() => {
                  // Error toast is already shown by useMutation.
                });
              }}
            />
          )}
        </div>
      </div>
    </Card>
  );
}
