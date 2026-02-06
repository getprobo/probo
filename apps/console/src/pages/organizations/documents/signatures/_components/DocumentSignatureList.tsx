import { useEffect, useRef } from "react";
import { useFragment, useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentSignatureList_peopleFragment$key } from "#/__generated__/core/DocumentSignatureList_peopleFragment.graphql";
import type { DocumentSignatureList_versionFragment$key } from "#/__generated__/core/DocumentSignatureList_versionFragment.graphql";
import type { DocumentSignatureListQuery } from "#/__generated__/core/DocumentSignatureListQuery.graphql";

import { DocumentSignatureListItem } from "./DocumentSignatureListItem";
import { DocumentSignaturePlaceholder } from "./DocumentSignaturePlaceholder";

const versionFragment = graphql`
  fragment DocumentSignatureList_versionFragment on DocumentVersion
  @refetchable(queryName: "DocumentSignatureListQuery")
  @argumentDefinitions(
    count: { type: "Int", defaultValue: 1000 }
    cursor: { type: "CursorKey" }
    signatureFilter: { type: "DocumentVersionSignatureFilter" }
  ) {
    ...DocumentSignaturePlaceholder_versionFragment
    signatures(first: $count, after: $cursor, filter: $signatureFilter)
      @connection(
        key: "DocumentSignaturesTab_signatures"
        filters: ["filter"]
      ) {
      __id
      edges {
        node {
          id
          signedBy {
            id
          }
          ...DocumentSignatureListItemFragment
        }
      }
    }
  }
`;

const peopleFragment = graphql`
  fragment DocumentSignatureList_peopleFragment on Organization  @argumentDefinitions(
    filter: { type: "ProfileFilter" }
  ) {
    ...DocumentSignaturePlaceholder_organizationFragment
    profiles(
      first: 1000
      orderBy: { direction: ASC, field: FULL_NAME }
      filter: $filter
    ) {
      edges {
        node {
          id
          ...DocumentSignaturePlaceholder_personFragment
        }
      }
    }
  }
`;

type SignatureState = "REQUESTED" | "SIGNED";

export function DocumentSignatureList(props: {
  peopleFragmentRef: DocumentSignatureList_peopleFragment$key;
  versionFragmentRef: DocumentSignatureList_versionFragment$key;
  selectedStates: SignatureState[];
}) {
  const { peopleFragmentRef, selectedStates, versionFragmentRef } = props;

  const { profiles, ...organization } = useFragment<DocumentSignatureList_peopleFragment$key>(
    peopleFragment,
    peopleFragmentRef,
  );
  const [version, refetch] = useRefetchableFragment<
    DocumentSignatureListQuery,
    DocumentSignatureList_versionFragment$key
  >(
    versionFragment,
    versionFragmentRef,
  );
  const signatureMap = new Map(version.signatures.edges.map(({ node }) => [node.signedBy.id, node]));

  const isFirstRender = useRef(true);

  // Refetch when filter changes (skip initial render)
  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false;
      return;
    }

    const filter
      = selectedStates.length > 0 ? { states: selectedStates } : null;

    refetch({ signatureFilter: filter });
  }, [selectedStates, refetch]);

  const filteredPeople
    = selectedStates.length > 0
      ? profiles.edges.filter(({ node }) => signatureMap.has(node.id))
      : profiles.edges;

  return (
    <div className="space-y-2 divide-y divide-border-solid">
      {filteredPeople.map(({ node: p }) => {
        const signature = signatureMap.get(p.id);
        return (
          signature
            ? (
                <DocumentSignatureListItem
                  key={signature.id}
                  fragmentRef={signature}
                  connectionId={version.signatures.__id}
                />
              )
            : (
                <DocumentSignaturePlaceholder
                  connectionId={version.signatures.__id}
                  key={p.id}
                  personFragmentRef={p}
                  organizationFragmentRef={organization}
                  versionFragmentRef={version}
                />
              )
        );
      })}
    </div>
  );
}
