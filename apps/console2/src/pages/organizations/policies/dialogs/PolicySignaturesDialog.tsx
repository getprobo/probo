import {
  Avatar,
  Badge,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogTitle,
  IconCircleCheck,
  IconClock,
  PolicyVersionBadge,
  Spinner,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useState, type ReactNode, Suspense } from "react";
import type { PolicyPagePolicyFragment$data } from "../__generated__/PolicyPagePolicyFragment.graphql";
import clsx from "clsx";
import type { ItemOf, NodeOf } from "/types";
import { graphql, useFragment } from "react-relay";
import type { PolicySignaturesDialog_version$key } from "./__generated__/PolicySignaturesDialog_version.graphql";
import type { PolicySignaturesDialog_signature$key } from "./__generated__/PolicySignaturesDialog_signature.graphql";
import { usePeople } from "/hooks/graph/PeopleGraph.ts";
import { useOrganizationId } from "/hooks/useOrganizationId.ts";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts.ts";
import { sprintf } from "@probo/helpers";

type Props = {
  policy: PolicyPagePolicyFragment$data;
  children?: ReactNode;
};

type Version = NodeOf<PolicyPagePolicyFragment$data["versions"]>;

export function PolicySignaturesDialog(props: Props) {
  const { __ } = useTranslate();
  const versions = props.policy.versions.edges.map((edge) => edge.node);
  const [selectedVersionId, setSelectedVersionId] = useState<string>(
    versions[0].id
  );
  const selectedVersion = versions.find((v) => v.id === selectedVersionId);

  if (!selectedVersion) {
    return null;
  }

  return (
    <Dialog trigger={props.children}>
      <DialogContent className="flex" scrollableChildren>
        <aside className="p-6 overflow-y-auto w-60 flex-none space-y-2">
          <div className="text-base text-txt-primary font-medium mb-4">
            {__("Version History")}
          </div>
          {versions.map((version) => (
            <VersionItem
              key={version.id}
              version={version}
              active={selectedVersionId === version.id}
              onSelect={setSelectedVersionId}
            />
          ))}
        </aside>
        <main className="flex-1 px-12 py-8">
          <DialogTitle className="text-2xl font-bold text-txt-primary mb-4">
            {__("Signatures")}
          </DialogTitle>
          <p className="text-sm text-txt-secondary mb-4">
            {__(
              "Click request to ask for a signature from people in your organization."
            )}
          </p>
          <Suspense fallback={<Spinner centered />}>
            <SignatureList version={selectedVersion} />
          </Suspense>
        </main>
      </DialogContent>
      <DialogFooter />
    </Dialog>
  );
}

const versionFragment = graphql`
  fragment PolicySignaturesDialog_version on PolicyVersion {
    version
    status
    publishedAt
    updatedAt
  }
`;

function VersionItem(props: {
  version: Version;
  active?: boolean;
  onSelect: (v: string) => void;
}) {
  const { dateTimeFormat, __ } = useTranslate();
  const version = useFragment<PolicySignaturesDialog_version$key>(
    versionFragment,
    props.version
  );
  return (
    <button
      onClick={() => props.onSelect(props.version.id)}
      className={clsx(
        "text-start space-y-1 w-full overflow-hidden py-2 px-3 w-full hover:bg-tertiary-hover cursor-pointer rounded",
        props.active && "bg-tertiary-pressed"
      )}
    >
      <div className="flex gap-1 text-sm text-txt-primary">
        <span>
          {__("Version")} {version.version}
        </span>
        <PolicyVersionBadge state={version.status} />
      </div>
      <div className="text-xs text-txt-secondary">
        {dateTimeFormat(version.publishedAt ?? version.updatedAt)}
      </div>
    </button>
  );
}

function SignatureList(props: { version: Version }) {
  const signatures = props.version.signatures.edges.map((edge) => edge.node);
  const { __ } = useTranslate();
  const signatureMap = new Map(signatures.map((s) => [s.signedBy.id, s]));
  const organizationId = useOrganizationId();
  const people = usePeople(organizationId);

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
        />
      ))}
    </div>
  );
}

/**
 * Fragments
 */
const signatureFragment = graphql`
  fragment PolicySignaturesDialog_signature on PolicyVersionSignature {
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
  mutation PolicySignaturesDialog_requestSignatureMutation(
    $input: RequestSignatureInput!
    $connections: [ID!]!
  ) {
    requestSignature(input: $input) {
      policyVersionSignatureEdge @prependEdge(connections: $connections) {
        node {
          id
          state
          signedBy {
            id
          }
          ...PolicySignaturesDialog_signature
        }
      }
    }
  }
`;

function SignatureItem(props: {
  versionId: string;
  signature?: PolicySignaturesDialog_signature$key;
  people: ItemOf<ReturnType<typeof usePeople>>;
  connectionId: string;
}) {
  const signature = useFragment(signatureFragment, props.signature);
  const { __, dateTimeFormat } = useTranslate();
  const [requestSignature, isSendingRequest] = useMutationWithToasts(
    requestSignatureMutation,
    {
      successMessage: __("Signature request sent successfully"),
      errorMessage: __("Failed to send signature request"),
    }
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
        <Button
          variant="secondary"
          className="ml-auto"
          disabled={isSendingRequest}
          onClick={() => {
            requestSignature({
              variables: {
                input: {
                  policyVersionId: props.versionId,
                  signatoryId: props.people.id,
                },
                connections: [props.connectionId],
              },
            });
          }}
        >
          {__("Request signature")}
        </Button>
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
                isSigned ? signature.signedAt : signature.requestedAt
              )
            )}
          </span>
        </div>
      </div>
      <Badge variant={isSigned ? "success" : "warning"} className="ml-auto">
        {isSigned ? __("Signed") : __("Pending")}
      </Badge>
    </div>
  );
}
