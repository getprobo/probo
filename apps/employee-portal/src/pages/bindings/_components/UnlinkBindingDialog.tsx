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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import { type ReactElement, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { UnlinkBindingDialogMutation } from "#/__generated__/core/UnlinkBindingDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const unlinkBindingDialogMutation = graphql`
  mutation UnlinkBindingDialogMutation($input: DeleteProbotIdentityBindingInput!) {
    deleteProbotIdentityBinding(input: $input) {
      probotIdentityBindingId
      viewer {
        probotIdentityBindings {
          id
          ...BindingListItem_binding
        }
      }
    }
  }
`;

export interface UnlinkBindingDialogProps {
  bindingId: string;
  children: ReactElement;
}

export function UnlinkBindingDialog({
  bindingId,
  children,
}: UnlinkBindingDialogProps) {
  const { t } = useTranslation("bindings");
  const [open, setOpen] = useState(false);
  const [unlinkBinding, isUnlinking]
    = useMutation<UnlinkBindingDialogMutation>(
      unlinkBindingDialogMutation,
      {
        successMessage: t("list.messages.unlinked"),
        errorToast: t("list.errors.unlinkFailed"),
      },
    );

  function handleUnlink() {
    void unlinkBinding({
      variables: { input: { id: bindingId } },
    }).then(
      () => {
        setOpen(false);
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={children} />
      <DialogPopup className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("list.confirmations.unlinkTitle")}</DialogTitle>
          <DialogDescription>
            {t("list.confirmations.unlink")}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral" highContrast>
                {t("list.actions.cancel")}
              </Button>
            )}
          />
          <Button
            variant="solid"
            color="red"
            loading={isUnlinking}
            onClick={handleUnlink}
          >
            {t("list.actions.unlink")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
