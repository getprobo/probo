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

import {
  CaretRightIcon,
  FileTextIcon,
  KeyIcon,
  SignOutIcon,
  UserIcon,
} from "@phosphor-icons/react";
import { formatError } from "@probo/helpers";
import { useToast } from "@probo/ui";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownGroup } from "@probo/ui/src/v2/Dropdown/DropdownGroup";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownSeparator } from "@probo/ui/src/v2/Dropdown/DropdownSeparator";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { EditableAvatarButton } from "@probo/ui/src/v2/EditableAvatarButton/EditableAvatarButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";
import { Link } from "react-router";

import type { ViewerMembershipMenu_organization$key } from "#/__generated__/iam/ViewerMembershipMenu_organization.graphql";
import type { ViewerMembershipMenuSignOutMutation } from "#/__generated__/iam/ViewerMembershipMenuSignOutMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { employeePortalHref } from "#/lib/employeePortalHref";
import { IdentityAvatarDialog } from "#/pages/iam/_components/IdentityAvatarDialog";

import { navRail } from "./variants";

const viewerMembershipMenuFragment = graphql`
  fragment ViewerMembershipMenu_organization on Organization {
    viewer @required(action: THROW) {
      fullName
      identity @required(action: THROW) {
        email
        avatar {
          downloadUrl
        }
        canListOAuth2AccessTokens: permission(
          action: "iam:oauth2-access-token:list"
        )
        ...IdentityAvatarDialog_identity
      }
    }
  }
`;

const signOutMutation = graphql`
  mutation ViewerMembershipMenuSignOutMutation {
    signOut {
      success
    }
  }
`;

export interface ViewerMembershipMenuProps {
  organizationKey: ViewerMembershipMenu_organization$key;
}

export function ViewerMembershipMenu({ organizationKey }: ViewerMembershipMenuProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { toast } = useToast();

  const {
    viewer: {
      fullName,
      identity,
    },
  } = useFragment(viewerMembershipMenuFragment, organizationKey);
  const { canListOAuth2AccessTokens, email, avatar } = identity;
  const [signOut, isSigningOut] = useMutation<ViewerMembershipMenuSignOutMutation>(signOutMutation);
  const [avatarOpen, setAvatarOpen] = useState(false);

  const displayName = fullName.trim() || email;
  const avatarSrc = avatar?.downloadUrl;

  const handleSignOut = () => {
    signOut({
      variables: {},
      onCompleted: (_, errors) => {
        if (errors) {
          toast({
            title: t("viewerMembershipDropdown.errors.requestFailed"),
            description: formatError(t("viewerMembershipDropdown.errors.cannotSignOut"), errors),
            variant: "error",
          });
          return;
        }
        // Full reload rather than a client navigation, so no Relay store
        // survives the session it belonged to.
        window.location.href = "/auth/login";
      },
      onError: (error) => {
        toast({
          title: t("common.error"),
          description: error.message,
          variant: "error",
        });
      },
    });
  };

  const railSlots = navRail();

  const trigger = (
    <button type="button" className={railSlots.item()}>
      <span className={railSlots.icon()}>
        <Avatar
          size={2}
          variant="soft"
          color="gold"
          radius="full"
          src={avatarSrc}
          fallback={displayName.charAt(0).toUpperCase() || <UserIcon />}
        />
      </span>
      <Text size={2} weight="medium" color="neutral" highContrast className={railSlots.label()}>
        {displayName}
      </Text>
      <CaretRightIcon className={railSlots.caret()} />
    </button>
  );

  return (
    <>
      <Dropdown>
        <DropdownTrigger render={trigger} />
        <DropdownPopup side="right" sideOffset={12} align="end">
          <DropdownGroup>
            <div className="flex w-full items-center gap-3 px-3 py-3">
              <EditableAvatarButton
                fullName={displayName}
                src={avatarSrc}
                fallback={displayName.charAt(0).toUpperCase() || <UserIcon />}
                onClick={() => setAvatarOpen(true)}
                label={t("editAvatar.actions.change")}
                size={2}
                radius="full"
              />
              <div className="flex min-w-0 flex-col gap-1">
                <Text size={2} weight="medium" color="neutral" highContrast>
                  {displayName}
                </Text>
                <Text size={1} color="faint" className="truncate">
                  {email}
                </Text>
              </div>
            </div>
          </DropdownGroup>
          <DropdownSeparator />
          {canListOAuth2AccessTokens && (
            <DropdownItem iconStart={<KeyIcon />} render={<Link to="/me/oauth-tokens" />}>
              {t("viewerMembershipDropdown.actions.oauthTokens")}
            </DropdownItem>
          )}
          <DropdownItem
            iconStart={<FileTextIcon />}
            render={<a href={employeePortalHref(organizationId)} />}
          >
            {t("viewerMembershipDropdown.actions.employeePortal")}
          </DropdownItem>
          <DropdownSeparator />
          <DropdownItem
            color="error"
            iconStart={<SignOutIcon />}
            disabled={isSigningOut}
            onClick={handleSignOut}
          >
            {t("viewerMembershipDropdown.actions.logout")}
          </DropdownItem>
        </DropdownPopup>
      </Dropdown>
      <IdentityAvatarDialog
        identityKey={identity}
        open={avatarOpen}
        onOpenChange={setAvatarOpen}
      />
    </>
  );
}
