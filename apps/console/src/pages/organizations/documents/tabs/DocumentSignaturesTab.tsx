import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Avatar,
  Badge,
  Button,
  Checkbox,
  IconCircleCheck,
  IconClock,
  Spinner,
} from "@probo/ui";
import { Suspense, useEffect, useRef, useState } from "react";
import { graphql, useFragment, useRefetchableFragment } from "react-relay";
import { useOutletContext } from "react-router";

import type { DocumentDetailPageDocumentFragment$data } from "#/__generated__/core/DocumentDetailPageDocumentFragment.graphql";
import type { DocumentSignaturesTab_signature$key } from "#/__generated__/core/DocumentSignaturesTab_signature.graphql.ts";
import type { DocumentSignaturesTab_version$key } from "#/__generated__/core/DocumentSignaturesTab_version.graphql.ts";
import type { DocumentSignaturesTabRefetchQuery } from "#/__generated__/core/DocumentSignaturesTabRefetchQuery.graphql";
import { usePeople } from "#/hooks/graph/PeopleGraph.ts";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts.ts";
import { useOrganizationId } from "#/hooks/useOrganizationId.ts";
import type { ItemOf, NodeOf } from "#/types";

type Version = NodeOf<DocumentDetailPageDocumentFragment$data["versions"]>;

const versionFragment = graphql`
  fragment DocumentSignaturesTab_version on DocumentVersion
  @refetchable(queryName: "DocumentSignaturesTabRefetchQuery")
  @argumentDefinitions(
    count: { type: "Int", defaultValue: 1000 }
    cursor: { type: "CursorKey" }
    signatureFilter: { type: "DocumentVersionSignatureFilter" }
  ) {
    id
    status
    canRequestSignature: permission(
      action: "core:document-version:request-signature"
    )
    signatures(first: $count, after: $cursor, filter: $signatureFilter)
      @connection(
        key: "DocumentSignaturesTab_signatures"
        filters: ["filter"]
      ) {
      __id
      edges {
        node {
          id
          state
          signedBy {
            id
            fullName
            primaryEmailAddress
          }
          ...DocumentSignaturesTab_signature
        }
      }
    }
  }
`;

type SignatureState = "REQUESTED" | "SIGNED";

export default function DocumentSignaturesTab() {
  const { version: versionFromContext } = useOutletContext<{
    version: Version;
  }>();
  const [selectedStates, setSelectedStates] = useState<SignatureState[]>([]);
  const { __ } = useTranslate();

  if (!versionFromContext) {
    return null;
  }

  const toggleState = (state: SignatureState) => {
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
            onChange={() => toggleState("REQUESTED")}
          />
          <span
            className="text-sm text-txt-secondary cursor-pointer select-none"
            onClick={() => toggleState("REQUESTED")}
          >
            {__("Requested")}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <Checkbox
            checked={selectedStates.includes("SIGNED")}
            onChange={() => toggleState("SIGNED")}
          />
          <span
            className="text-sm text-txt-secondary cursor-pointer select-none"
            onClick={() => toggleState("SIGNED")}
          >
            {__("Signed")}
          </span>
        </div>
      </div>
      <Suspense fallback={<Spinner centered />}>
        <SignatureList
          version={versionFromContext}
          selectedStates={selectedStates}
        />
      </Suspense>
    </div>
  );
}

function SignatureList(props: {
  version: Version;
  selectedStates: SignatureState[];
}) {
  const { version: propVersion, selectedStates } = props;
  const [version, refetch] = useRefetchableFragment<
    DocumentSignaturesTabRefetchQuery,
    DocumentSignaturesTab_version$key
  >(versionFragment, propVersion);
  const signatures = version.signatures?.edges?.map(edge => edge.node) ?? [];
  const { __ } = useTranslate();
  const signatureMap = new Map(signatures.map(s => [s.signedBy.id, s]));
  const organizationId = useOrganizationId();
  const people = usePeople(organizationId, { excludeContractEnded: true });
  const signable = version.status === "PUBLISHED";
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

  if (!version.signatures) {
    return (
      <div className="text-center text-sm text-txt-tertiary py-3">
        {__("Loading signatures...")}
      </div>
    );
  }

  if (people.length === 0) {
    return (
      <div className="text-center text-sm text-txt-tertiary py-3">
        {__("No people available to request signatures from")}
      </div>
    );
  }

  // When a filter is active, only show people who have signatures in the filtered results
  const filteredPeople
    = selectedStates.length > 0
      ? people.filter(p => signatureMap.has(p.id))
      : people;

  if (filteredPeople.length === 0 && selectedStates.length > 0) {
    return (
      <div className="text-center text-sm text-txt-tertiary py-3">
        {__("No signatures match the selected filters")}
      </div>
    );
  }

  return (
    <div className="space-y-2 divide-y divide-border-solid">
      {filteredPeople.map(p => (
        <SignatureItem
          key={p.id}
          versionId={version.id}
          signature={signatureMap.get(p.id)}
          people={p}
          connectionId={version.signatures.__id}
          signable={signable}
          canRequestSignature={version.canRequestSignature}
        />
      ))}
    </div>
  );
}

const signatureFragment = graphql`
  fragment DocumentSignaturesTab_signature on DocumentVersionSignature {
    id
    state
    signedAt
    requestedAt
    signedBy {
      fullName
      primaryEmailAddress
    }
    canCancel: permission(action: "core:document-version:request-signature")
  }
`;

/**
 * Mutations
 */
const requestSignatureMutation = graphql`
  mutation DocumentSignaturesTab_requestSignatureMutation(
    $input: RequestSignatureInput!
    $connections: [ID!]!
  ) {
    requestSignature(input: $input) {
      documentVersionSignatureEdge @prependEdge(connections: $connections) {
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
`;
const cancelSignatureMutation = graphql`
  mutation DocumentSignaturesTab_cancelSignatureMutation(
    $input: CancelSignatureRequestInput!
    $connections: [ID!]!
  ) {
    cancelSignatureRequest(input: $input) {
      deletedDocumentVersionSignatureId @deleteEdge(connections: $connections)
    }
  }
`;

function SignatureItem(props: {
  versionId: string;
  signature?: DocumentSignaturesTab_signature$key;
  people: ItemOf<ReturnType<typeof usePeople>>;
  connectionId: string;
  signable: boolean;
  canRequestSignature: boolean;
}) {
  const signature = useFragment(signatureFragment, props.signature);
  const { __, dateTimeFormat } = useTranslate();
  const [requestSignature, isSendingRequest] = useMutationWithToasts(
    requestSignatureMutation,
    {
      successMessage: __("Signature request sent successfully"),
      errorMessage: __("Failed to send signature request"),
    },
  );
  const [cancelSignature, isCancellingSignature] = useMutationWithToasts(
    cancelSignatureMutation,
    {
      successMessage: __("Request cancelled successfully"),
      errorMessage: __("Failed to cancel signature request"),
    },
  );

  // No signature request for this user
  if (!signature) {
    return (
      <div className="flex gap-3 items-center py-3">
        <Avatar size="l" name={props.people.fullName} />
        <div className="space-y-1">
          <div className="text-sm text-txt-primary font-medium">
            {props.people.fullName}
          </div>
          <div className="text-xs text-txt-secondary">
            {props.people.primaryEmailAddress}
          </div>
        </div>
        {props.signable && props.canRequestSignature && (
          <Button
            variant="secondary"
            className="ml-auto"
            disabled={isSendingRequest}
            onClick={() => {
              void requestSignature({
                variables: {
                  input: {
                    documentVersionId: props.versionId,
                    signatoryId: props.people.id,
                  },
                  connections: [props.connectionId],
                },
              });
            }}
          >
            {__("Request signature")}
          </Button>
        )}
      </div>
    );
  }

  const isSigned = signature.state === "SIGNED";
  const label = isSigned ? __("Signed on %s") : __("Requested on %s");

  return (
    <div className="flex gap-3 items-center py-3">
      <Avatar size="l" name={signature.signedBy.fullName} />
      <div className="space-y-1">
        <div className="text-sm text-txt-primary font-medium">
          {signature.signedBy.fullName}
        </div>
        <div className="text-xs text-txt-secondary flex items-center gap-1">
          {isSigned
            ? (
                <IconCircleCheck size={16} className="text-txt-accent" />
              )
            : (
                <IconClock size={16} />
              )}
          <span>
            {sprintf(
              label,
              dateTimeFormat(
                isSigned ? signature.signedAt : signature.requestedAt,
              ),
            )}
          </span>
        </div>
      </div>
      {isSigned
        ? (
            <Badge variant="success" className="ml-auto">
              {__("Signed")}
            </Badge>
          )
        : (
            signature.canCancel && (
              <Button
                variant="danger"
                className="ml-auto"
                disabled={isCancellingSignature}
                onClick={() => {
                  void cancelSignature({
                    variables: {
                      input: {
                        documentVersionSignatureId: signature.id,
                      },
                      connections: [props.connectionId],
                    },
                  });
                }}
              >
                {__("Cancel request")}
              </Button>
            )
          )}
    </div>
  );
}
