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

import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Button, IconCircleCheck, IconClock } from "@probo/ui";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { DocumentSignatureListItemFragment$key } from "#/__generated__/core/DocumentSignatureListItemFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const fragment = graphql`
  fragment DocumentSignatureListItemFragment on DocumentVersionSignature {
    id
    signedBy {
      fullName
    }
    state
    signedAt
    requestedAt
    canCancel: permission(action: "core:document-version-signature:cancel")
  }
`;

const cancelSignatureMutation = graphql`
  mutation DocumentSignatureListItem_cancelSignatureMutation(
    $input: CancelSignatureRequestInput!
    $connections: [ID!]!
  ) {
    cancelSignatureRequest(input: $input) {
      deletedDocumentVersionSignatureId @deleteEdge(connections: $connections)
    }
  }
`;

export function DocumentSignatureListItem(props: {
  fragmentRef: DocumentSignatureListItemFragment$key;
  connectionId: DataID;
}) {
  const { connectionId, fragmentRef } = props;

  const { __, dateTimeFormat } = useTranslate();
  const signature = useFragment<DocumentSignatureListItemFragment$key>(fragment, fragmentRef);

  const isSigned = signature.state === "SIGNED";
  const label = isSigned ? __("Signed on %s") : __("Requested on %s");

  const [cancelSignature, isCancellingSignature] = useMutationWithToasts(
    cancelSignatureMutation,
    {
      successMessage: __("Request cancelled successfully"),
      errorMessage: __("Failed to cancel signature request"),
    },
  );

  return (
    <div className="flex gap-3 items-center py-3">
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
                      connections: [connectionId],
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
