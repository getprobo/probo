import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Avatar, Badge, Button, IconCircleCheck, IconClock } from "@probo/ui";
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
