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

import { Checkbox, Spinner } from "@probo/ui";
import { Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentSignatureList_versionFragment$key } from "#/__generated__/core/DocumentSignatureList_versionFragment.graphql";
import type { DocumentSignaturesPageQuery } from "#/__generated__/core/DocumentSignaturesPageQuery.graphql";

import { DocumentSignatureList } from "./_components/DocumentSignatureList";

export const documentSignaturesPageQuery = graphql`
  query DocumentSignaturesPageQuery($documentId: ID! $organizationId: ID! $versionId: ID! $versionSpecified: Boolean!) {
    organization: node(id: $organizationId) {
      __typename
      ...DocumentSignatureList_peopleFragment @arguments(filter: { contractEnded: false, state: ACTIVE })
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

  const { t } = useTranslation();

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
          {t("documentSignaturesPage.filter.label")}
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
            {t("documentSignaturesPage.states.requested")}
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
            {t("documentSignaturesPage.states.signed")}
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
