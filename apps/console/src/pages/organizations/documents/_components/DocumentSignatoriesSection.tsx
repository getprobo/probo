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

import { MagnifyingGlassIcon, SignatureIcon, XIcon } from "@phosphor-icons/react";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, usePaginationFragment } from "react-relay";

import type { DocumentSignatoriesSection_cancelMutation } from "#/__generated__/core/DocumentSignatoriesSection_cancelMutation.graphql";
import type { DocumentSignatoriesSection_document$key } from "#/__generated__/core/DocumentSignatoriesSection_document.graphql";
import type { DocumentSignatoriesSectionList_version$key } from "#/__generated__/core/DocumentSignatoriesSectionList_version.graphql";
import type { DocumentSignatoriesSectionListRefetchQuery } from "#/__generated__/core/DocumentSignatoriesSectionListRefetchQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { cardListPageSize, useCardListSearch } from "../_lib/useCardListSearch";
import { documentSignatories } from "../variants";

import { DocumentSignatoryMenu } from "./DocumentSignatoryMenu";

// Must match the default filter of the signatures connection below.
const activeSignatoriesFilter = { activeContract: true, profileStates: ["ACTIVE" as const] };

const documentSignatoriesSectionFragment = graphql`
  fragment DocumentSignatoriesSection_document on Document {
    publishedVersion: versions(
      first: 1
      filter: { statuses: [PUBLISHED] }
      orderBy: { field: CREATED_AT, direction: DESC }
    ) {
      edges {
        node {
          id
          ...DocumentSignatoriesSectionList_version
        }
      }
    }
  }
`;

const documentSignatoriesListFragment = graphql`
  fragment DocumentSignatoriesSectionList_version on DocumentVersion
  @refetchable(queryName: "DocumentSignatoriesSectionListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
    filter: {
      type: "DocumentVersionSignatureFilter"
      defaultValue: { activeContract: true, profileStates: [ACTIVE] }
    }
  ) {
    id
    canRequestSignature: permission(action: "core:document-version-signature:request")
    allSignatures: signatures(first: 0, filter: { activeContract: true, profileStates: [ACTIVE] }) {
      totalCount
    }
    signatures(
      first: $first
      after: $after
      filter: $filter
    ) @connection(key: "DocumentSignatoriesSection_signatures") {
      __id
      edges {
        node {
          id
          state
          canCancel: permission(action: "core:document-version-signature:cancel")
          signedBy {
            id
            fullName
          }
        }
      }
    }
  }
`;

const cancelSignatureMutation = graphql`
  mutation DocumentSignatoriesSection_cancelMutation(
    $input: CancelSignatureRequestInput!
    $connections: [ID!]!
  ) {
    cancelSignatureRequest(input: $input) {
      deletedDocumentVersionSignatureId @deleteEdge(connections: $connections)
    }
  }
`;

interface DocumentSignatoriesSectionProps {
  documentKey: DocumentSignatoriesSection_document$key;
  // A previous version on screen, in place of the published one.
  versionKey?: DocumentSignatoriesSectionList_version$key | null;
}

function stateColor(state: "REQUESTED" | "SIGNED") {
  switch (state) {
    case "SIGNED":
      return "green";
    case "REQUESTED":
      return "amber";
  }
}

export function DocumentSignatoriesSection({ documentKey, versionKey }: DocumentSignatoriesSectionProps) {
  const document = useFragment(documentSignatoriesSectionFragment, documentKey);
  const published = document.publishedVersion.edges[0]?.node ?? null;
  const version = versionKey ?? published;
  const { root, title, titleLead, titleIcon } = documentSignatories();
  const { t } = useTranslation("organizations/documents");

  if (version == null) {
    return (
      <Card variant="soft" size={2}>
        <div className={root()}>
          <div className={title()}>
            <div className={titleLead()}>
              <SignatureIcon className={titleIcon()} aria-hidden />
              <Heading level={2} size={2} weight="medium">
                {t("details.signatories.title")}
              </Heading>
            </div>
          </div>
          <Text size={1} color="faint">{t("details.signatories.none")}</Text>
        </div>
      </Card>
    );
  }

  return <DocumentSignatoriesList versionKey={version} publishedVersionId={published?.id ?? null} />;
}

function DocumentSignatoriesList({
  versionKey,
  publishedVersionId,
}: {
  versionKey: DocumentSignatoriesSectionList_version$key;
  // Signatures are only requested or cancelled on the published version.
  publishedVersionId: string | null;
}) {
  const { t } = useTranslation("organizations/documents");
  const {
    data: version,
    loadNext,
    hasNext,
    isLoadingNext,
    refetch,
  } = usePaginationFragment<
    DocumentSignatoriesSectionListRefetchQuery,
    DocumentSignatoriesSectionList_version$key
  >(documentSignatoriesListFragment, versionKey);
  const { query, setQuery, filtering, isPending } = useCardListSearch((next) => {
    refetch({
      first: cardListPageSize,
      filter: { ...activeSignatoriesFilter, query: next === "" ? null : next },
    }, { fetchPolicy: "network-only" });
  });
  const [cancelSignature, isCancelling] = useMutation<DocumentSignatoriesSection_cancelMutation>(
    cancelSignatureMutation,
    { errorToast: t("details.errors.update") },
  );
  const { root, title, titleLead, titleIcon, rows, row, name, trailing } = documentSignatories({
    pending: isPending,
  });
  const signatures = version.signatures.edges.map(edge => edge.node);
  const signatoryIds = new Set(signatures.map(signature => signature.signedBy.id));
  const editable = version.id === publishedVersionId;

  return (
    <Card variant="soft" size={2}>
      <div className={root()}>
        <div className={title()}>
          <div className={titleLead()}>
            <SignatureIcon className={titleIcon()} aria-hidden />
            <Heading level={2} size={2} weight="medium">
              {t("details.signatories.title")}
            </Heading>
          </div>
          {editable && version.canRequestSignature && (
            <DocumentSignatoryMenu
              documentVersionId={version.id}
              connectionId={version.signatures.__id}
              signatoryIds={signatoryIds}
              disabled={isCancelling || isLoadingNext}
            />
          )}
        </div>
        {version.allSignatures.totalCount > cardListPageSize && (
          <TextField
            size={1}
            icon={<MagnifyingGlassIcon />}
            value={query}
            onValueChange={setQuery}
            placeholder={t("details.signatories.filter")}
            aria-label={t("details.signatories.filter")}
          />
        )}
        <div className={rows()} aria-busy={isPending}>
          {signatures.length > 0
            ? signatures.map(signature => (
                <div key={signature.id} className={row()}>
                  <Text size={2} className={name()} title={signature.signedBy.fullName}>
                    {signature.signedBy.fullName}
                  </Text>
                  <div className={trailing()}>
                    <Badge size={1} variant="soft" color={stateColor(signature.state)}>
                      {t(`details.signatories.states.${signature.state}`)}
                    </Badge>
                    {editable && signature.state === "REQUESTED" && signature.canCancel && (
                      <IconButton
                        size={1}
                        variant="ghost"
                        color="neutral"
                        disabled={isCancelling}
                        aria-label={t("details.signatories.remove", { name: signature.signedBy.fullName })}
                        onClick={() => {
                          void cancelSignature({
                            variables: {
                              input: { documentVersionSignatureId: signature.id },
                              connections: [version.signatures.__id],
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
                </div>
              ))
            : (
                <Text size={1} color="faint">
                  {filtering ? t("details.signatories.filterEmpty") : t("details.signatories.none")}
                </Text>
              )}
          {hasNext && (
            <Button
              variant="ghost"
              color="neutral"
              size={1}
              className="self-start"
              loading={isLoadingNext}
              onClick={() => loadNext(cardListPageSize)}
            >
              {t("details.signatories.loadMore")}
            </Button>
          )}
        </div>
      </div>
    </Card>
  );
}
