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

import { Button, Dialog, DialogContent, DialogFooter, type DialogRef, IconSend, Spinner } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { SendUpdateDialogMutation } from "#/__generated__/core/SendUpdateDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

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
  const { t } = useTranslation("organizations/compliance-page");

  const [sendUpdate, isSending] = useMutation<SendUpdateDialogMutation>(sendMutation, {
    successMessage: t("sendUpdateDialog.messages.enqueued"),
    errorToast: t("sendUpdateDialog.errors.enqueue"),
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
    <Dialog ref={ref} title={t("sendUpdateDialog.title")}>
      <DialogContent className="px-6 pt-5 pb-2 space-y-4">
        <p className="text-sm text-txt-secondary">
          {t("sendUpdateDialog.description")}
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
          {t("sendUpdateDialog.actions.send")}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
