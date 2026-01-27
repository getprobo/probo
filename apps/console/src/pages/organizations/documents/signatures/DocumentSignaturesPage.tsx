import { useTranslate } from "@probo/i18n";
import { Checkbox, Spinner } from "@probo/ui";
import { Suspense, useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentSignatureList_versionFragment$key } from "#/__generated__/core/DocumentSignatureList_versionFragment.graphql";
import type { DocumentSignaturesPageQuery } from "#/__generated__/core/DocumentSignaturesPageQuery.graphql";

import { DocumentSignatureList } from "./_components/DocumentSignatureList";

export const documentSignaturesPageQuery = graphql`
  query DocumentSignaturesPageQuery($documentId: ID! $organizationId: ID! $versionId: ID! $versionSpecified: Boolean!) {
    organization: node(id: $organizationId) {
      __typename
      ...DocumentSignatureList_peopleFragment @arguments(filter: { excludeContractEnded: true })
    }
    # We use this on /documents/:documentId
    document: node(id: $documentId) @skip(if: $versionSpecified) {
      __typename
      ... on Document {
        lastVersion: versions(
          first: 1
          orderBy: { field: CREATED_AT, direction: DESC }
        ) {
          edges {
            node {
              ...DocumentSignatureList_versionFragment
            }
          }
        }
      }
    }
    # We use this on /documents/:documentId/versions/:versionId
    version: node(id: $versionId) @include(if: $versionSpecified) {
      __typename
      ...DocumentSignatureList_versionFragment
    }
  }
`;

type SignatureState = "REQUESTED" | "SIGNED";

export function DocumentSignaturesPage(props: { queryRef: PreloadedQuery<DocumentSignaturesPageQuery> }) {
  const { queryRef } = props;

  const { __ } = useTranslate();

  const {
    organization,
    document,
    version,
  } = usePreloadedQuery<DocumentSignaturesPageQuery>(documentSignaturesPageQuery, queryRef);
  if (organization.__typename != "Organization" || (version && version.__typename != "DocumentVersion") || (document && document.__typename !== "Document")) {
    throw new Error("invalid type for node");
  }
  if (!document && !version) {
    throw new Error("no document or version sepcified");
  }

  const [selectedStates, setSelectedStates] = useState<SignatureState[]>([]);
  const handleSelectState = (state: SignatureState) => {
    setSelectedStates(prev =>
      prev.includes(state) ? prev.filter(s => s !== state) : [...prev, state],
    );
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4 pb-2 border-b border-border-solid">
        <span className="text-sm text-txt-secondary">
          {__("Filter by state:")}
        </span>
        <div className="flex items-center gap-2">
          <Checkbox
            checked={selectedStates.includes("REQUESTED")}
            onChange={() => handleSelectState("REQUESTED")}
          />
          <span
            className="text-sm text-txt-secondary cursor-pointer select-none"
            onClick={() => handleSelectState("REQUESTED")}
          >
            {__("Requested")}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <Checkbox
            checked={selectedStates.includes("SIGNED")}
            onChange={() => handleSelectState("SIGNED")}
          />
          <span
            className="text-sm text-txt-secondary cursor-pointer select-none"
            onClick={() => handleSelectState("SIGNED")}
          >
            {__("Signed")}
          </span>
        </div>
      </div>
      <Suspense fallback={<Spinner centered />}>
        <DocumentSignatureList
          peopleFragmentRef={organization}
          versionFragmentRef={
            (version ?? document?.lastVersion.edges[0].node) as DocumentSignatureList_versionFragment$key
          }
          selectedStates={selectedStates}
        />
      </Suspense>
    </div>
  );
}
