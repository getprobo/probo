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
    major
    minor
    document {
      currentPublishedMajor
      currentPublishedMinor
    }
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

  const isCurrentPublishedVersion
    = version.status === "PUBLISHED"
      && version.document.currentPublishedMajor === version.major
      && version.document.currentPublishedMinor === version.minor;

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
      {isCurrentPublishedVersion && canRequestSignature && (
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
