// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import { Button } from "@probo/ui";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { DocumentSignaturePlaceholder_organizationFragment$key } from "#/__generated__/core/DocumentSignaturePlaceholder_organizationFragment.graphql";
import type { DocumentSignaturePlaceholder_personFragment$key } from "#/__generated__/core/DocumentSignaturePlaceholder_personFragment.graphql";
import type { DocumentSignaturePlaceholder_versionFragment$key } from "#/__generated__/core/DocumentSignaturePlaceholder_versionFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const organizationFragment = graphql`
  fragment DocumentSignaturePlaceholder_organizationFragment on Organization {
    canRequestSignature: permission(
      action: "core:document-version:request-signature"
    )
  }
`;

const personFragment = graphql`
  fragment DocumentSignaturePlaceholder_personFragment on Profile {
    id
    fullName
    emailAddress
  }
`;

const versionFragment = graphql`
  fragment DocumentSignaturePlaceholder_versionFragment on DocumentVersion {
    id
    status
  }
`;

const requestSignatureMutation = graphql`
  mutation DocumentSignaturePlaceholder_requestSignatureMutation(
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
        }
      }
    }
  }
`;

export function DocumentSignaturePlaceholder(props: {
  connectionId: DataID;
  organizationFragmentRef: DocumentSignaturePlaceholder_organizationFragment$key;
  personFragmentRef: DocumentSignaturePlaceholder_personFragment$key;
  versionFragmentRef: DocumentSignaturePlaceholder_versionFragment$key;
}) {
  const { connectionId, organizationFragmentRef, personFragmentRef, versionFragmentRef } = props;

  const { __ } = useTranslate();

  const { canRequestSignature } = useFragment<DocumentSignaturePlaceholder_organizationFragment$key>(
    organizationFragment,
    organizationFragmentRef,
  );
  const person = useFragment<DocumentSignaturePlaceholder_personFragment$key>(personFragment, personFragmentRef);
  const version = useFragment<DocumentSignaturePlaceholder_versionFragment$key>(versionFragment, versionFragmentRef);

  const [requestSignature, isSendingRequest] = useMutationWithToasts(
    requestSignatureMutation,
    {
      successMessage: __("Signature request sent successfully"),
      errorMessage: __("Failed to send signature request"),
    },
  );

  return (
    <div className="flex gap-3 items-center py-3">
      <div className="space-y-1">
        <div className="text-sm text-txt-primary font-medium">
          {person.fullName}
        </div>
        <div className="text-xs text-txt-secondary">
          {person.emailAddress}
        </div>
      </div>
      {version.status === "PUBLISHED" && canRequestSignature && (
        <Button
          variant="secondary"
          className="ml-auto"
          disabled={isSendingRequest}
          onClick={() => {
            void requestSignature({
              variables: {
                input: {
                  documentVersionId: version.id,
                  signatoryId: person.id,
                },
                connections: [connectionId],
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
