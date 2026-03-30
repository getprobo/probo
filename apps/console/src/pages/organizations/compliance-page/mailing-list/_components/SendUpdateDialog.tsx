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
import { Button, Dialog, DialogContent, DialogFooter, type DialogRef, IconSend, Spinner } from "@probo/ui";
import { graphql } from "relay-runtime";

import type { SendUpdateDialogMutation } from "#/__generated__/core/SendUpdateDialogMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import type { UpdateNode } from "./CompliancePageUpdatesList";

const sendMutation = graphql`
  mutation SendUpdateDialogMutation($input: SendMailingListUpdateInput!) {
    sendMailingListUpdate(input: $input) {
      mailingListUpdate {
        id
        title
        body
        status
        updatedAt
      }
    }
  }
`;

type Props = {
  ref: DialogRef;
  update: UpdateNode | null;
  onSent?: () => void;
};

export function SendUpdateDialog({ ref, update, onSent }: Props) {
  const { __ } = useTranslate();

  const [sendUpdate, isSending] = useMutationWithToasts<SendUpdateDialogMutation>(sendMutation, {
    successMessage: __("Update enqueued for delivery"),
    errorMessage: __("Failed to enqueue update for delivery"),
  });

  const handleSend = async () => {
    if (!update) return;
    await sendUpdate({
      variables: { input: { id: update.id } },
      onCompleted: (_, errors) => {
        if (!errors?.length) {
          ref.current?.close();
          onSent?.();
        }
      },
    });
  };

  return (
    <Dialog ref={ref} title={__("Send Update to Subscribers")}>
      <DialogContent className="px-6 pt-5 pb-2 space-y-4">
        <p className="text-sm text-txt-secondary">
          {__("This update will be sent to all current subscribers of this mailing list. This action cannot be undone.")}
        </p>
        {update && (
          <div className="rounded-lg border border-border-low bg-surface-secondary overflow-hidden">
            <div className="px-4 py-3 border-b border-border-low">
              <p className="text-sm font-medium text-txt-primary">{update.title}</p>
            </div>
            <div className="px-4 py-3">
              <p className="text-sm text-txt-secondary whitespace-pre-wrap">{update.body}</p>
            </div>
          </div>
        )}
      </DialogContent>
      <DialogFooter>
        <Button
          icon={IconSend}
          disabled={isSending}
          onClick={() => void handleSend()}
        >
          {isSending && <Spinner />}
          {__("Send")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
