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

import { UserMinusIcon } from "@phosphor-icons/react";
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
import { graphql } from "relay-runtime";

import type { DeactivateVisitorDialogMutation } from "#/__generated__/core/DeactivateVisitorDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const deactivateAccessMutation = graphql`
  mutation DeactivateVisitorDialogMutation($input: DeactivateCompliancePortalAccessInput!) {
    deactivateCompliancePortalAccess(input: $input) {
      compliancePortalAccess {
        id
        state
      }
    }
  }
`;

interface DeactivateVisitorDialogProps {
  accessId: string;
  children: ReactElement;
}

export function DeactivateVisitorDialog({
  accessId,
  children,
}: DeactivateVisitorDialogProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const [open, setOpen] = useState(false);
  const [deactivateAccess, isDeactivating] = useMutation<DeactivateVisitorDialogMutation>(
    deactivateAccessMutation,
    {
      successMessage: t("visitorPage.messages.deactivated"),
      errorToast: t("visitorPage.errors.deactivate"),
    },
  );

  function handleDeactivate() {
    void deactivateAccess({
      variables: { input: { id: accessId } },
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
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("visitorPage.deactivate.title")}</DialogTitle>
          <DialogDescription>
            {t("visitorPage.deactivate.description")}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("visitorPage.deactivate.actions.cancel")}
              </Button>
            )}
          />
          <Button
            type="button"
            variant="solid"
            color="red"
            iconStart={<UserMinusIcon />}
            loading={isDeactivating}
            onClick={handleDeactivate}
          >
            {t("visitorPage.deactivate.actions.confirm")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
