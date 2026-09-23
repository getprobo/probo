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

import { ArchiveIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DeactivateUserDialog_deactivateMutation } from "#/__generated__/iam/DeactivateUserDialog_deactivateMutation.graphql";
import type { DeactivateUserDialog_profile$key } from "#/__generated__/iam/DeactivateUserDialog_profile.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const fragment = graphql`
  fragment DeactivateUserDialog_profile on Profile {
    id
    fullName
  }
`;

const deactivateUserMutation = graphql`
  mutation DeactivateUserDialog_deactivateMutation($input: DeactivateUserInput!) {
    deactivateUser(input: $input) {
      success
    }
  }
`;

interface DeactivateUserDialogProps {
  profileKey: DeactivateUserDialog_profile$key;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onDeactivated: () => void;
}

export function DeactivateUserDialog({
  profileKey,
  open,
  onOpenChange,
  onDeactivated,
}: DeactivateUserDialogProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const profile = useFragment(fragment, profileKey);
  const [deactivateUser, isDeactivating] = useMutation<DeactivateUserDialog_deactivateMutation>(
    deactivateUserMutation,
    {
      successMessage: t("userListItem.messages.deactivated"),
      errorToast: t("userListItem.errors.deactivate"),
    },
  );

  function handleDeactivate() {
    void deactivateUser({
      variables: {
        input: {
          profileId: profile.id,
          organizationId,
        },
      },
    }).then(
      () => {
        onOpenChange(false);
        onDeactivated();
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("userListItem.dialogs.deactivateTitle")}</DialogTitle>
          <DialogDescription>
            {t("userListItem.confirmations.deactivate", { name: profile.fullName })}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("usersPage.actions.cancel")}
              </Button>
            )}
          />
          <Button
            variant="solid"
            color="neutral"
            highContrast
            loading={isDeactivating}
            iconStart={<ArchiveIcon />}
            onClick={handleDeactivate}
          >
            {t("userListItem.actions.deactivatePerson")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
