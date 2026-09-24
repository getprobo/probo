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

import { EnvelopeIcon } from "@phosphor-icons/react";
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

import type { SendActivationEmailDialog_inviteMutation } from "#/__generated__/iam/SendActivationEmailDialog_inviteMutation.graphql";
import type { SendActivationEmailDialog_profile$key } from "#/__generated__/iam/SendActivationEmailDialog_profile.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const fragment = graphql`
  fragment SendActivationEmailDialog_profile on Profile {
    id
    fullName
    lastInvitation: pendingInvitations(first: 1, orderBy: { field: CREATED_AT, direction: DESC })
    @required(action: THROW)
    @connection(key: "SendActivationEmailDialog_lastInvitation") {
      __id
      # eslint-disable-next-line relay/unused-fields -- required by @connection
      edges {
        __typename
      }
    }
  }
`;

const inviteUserMutation = graphql`
  mutation SendActivationEmailDialog_inviteMutation(
    $input: InviteUserInput!
    $connections: [ID!]!
  ) {
    inviteUser(input: $input) {
      invitationEdge @prependEdge(connections: $connections) {
        node {
          id
          expiresAt
          acceptedAt
          createdAt
        }
      }
    }
  }
`;

interface SendActivationEmailDialogProps {
  profileKey: SendActivationEmailDialog_profile$key;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function SendActivationEmailDialog({
  profileKey,
  open,
  onOpenChange,
}: SendActivationEmailDialogProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const profile = useFragment(fragment, profileKey);
  const [inviteUser, isInviting] = useMutation<SendActivationEmailDialog_inviteMutation>(
    inviteUserMutation,
    {
      successMessage: t("userListItem.messages.invitationSent"),
      errorToast: t("userListItem.errors.sendInvitation"),
    },
  );

  function handleSend() {
    void inviteUser({
      variables: {
        input: {
          organizationId,
          profileId: profile.id,
        },
        connections: [profile.lastInvitation.__id],
      },
      updater: (store) => {
        store.get(profile.id)?.setValue("PENDING", "state");
      },
    }).then(
      () => {
        onOpenChange(false);
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
          <DialogTitle>{t("userListItem.dialogs.sendTitle")}</DialogTitle>
          <DialogDescription>
            {t("userListItem.confirmations.sendActivationEmail", { name: profile.fullName })}
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
            color="gold"
            loading={isInviting}
            iconStart={<EnvelopeIcon />}
            onClick={handleSend}
          >
            {t("userListItem.actions.send")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
