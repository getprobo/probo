import {
  Avatar,
  Badge,
  Button,
  IconCircleCheck,
  IconClock,
  Spinner,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { Suspense } from "react";
import type { ItemOf, NodeOf } from "/types";
import { graphql, useFragment } from "react-relay";
import { usePeople } from "/hooks/graph/PeopleGraph.ts";
import { useOrganizationId } from "/hooks/useOrganizationId.ts";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts.ts";
import { sprintf } from "@probo/helpers";
import type { DocumentDetailPageDocumentFragment$data } from "../__generated__/DocumentDetailPageDocumentFragment.graphql";
import { useOutletContext } from "react-router";
import type { DocumentSignaturesTab_signature$key } from "/pages/organizations/documents/tabs/__generated__/DocumentSignaturesTab_signature.graphql.ts";

type Version = NodeOf<DocumentDetailPageDocumentFragment$data["versions"]>;

export default function DocumentSignaturesTab() {
  const { version } = useOutletContext<{ version: Version }>();
  if (!version) {
    return null;
  }

  return (
    <Suspense fallback={<Spinner centered />}>
      <SignatureList version={version} />
    </Suspense>
  );
}

function SignatureList(props: { version: Version }) {
  const signatures = props.version.signatures?.edges?.map((edge) => edge.node) ?? [];
  const { __ } = useTranslate();
  const signatureMap = new Map(signatures.map((s) => [s.signedBy.id, s]));
  const organizationId = useOrganizationId();
  const people = usePeople(organizationId, { excludeContractEnded: true });
  const signable = props.version.status === "PUBLISHED";

  if (!props.version.signatures) {
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

  return (
    <div className="space-y-2 divide-y divide-border-solid">
      {people.map((p) => (
        <SignatureItem
          key={p.id}
          versionId={props.version.id}
          signature={signatureMap.get(p.id)}
          people={p}
          connectionId={props.version.signatures.__id}
          signable={signable}
        />
      ))}
    </div>
  );
}

/**
 * Fragments
 */
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
        {props.signable && (
          <Button
            variant="secondary"
            className="ml-auto"
            disabled={isSendingRequest}
            onClick={() => {
              requestSignature({
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
          {isSigned ? (
            <IconCircleCheck size={16} className="text-txt-accent" />
          ) : (
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
      {isSigned ? (
        <Badge variant="success" className="ml-auto">
          {__("Signed")}
        </Badge>
      ) : (
        <Button
          variant="danger"
          className="ml-auto"
          disabled={isCancellingSignature}
          onClick={() => {
            cancelSignature({
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
      )}
    </div>
  );
}
