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

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  DropdownSeparator,
  IconArrowBoxLeft,
  IconCircleQuestionmark,
  IconKey,
  IconPageTextLine,
  UserDropdown,
  UserDropdownItem,
  useToast,
} from "@probo/ui";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { ViewerMembershipDropdownFragment$key } from "#/__generated__/iam/ViewerMembershipDropdownFragment.graphql";
import type { ViewerMembershipDropdownSignOutMutation } from "#/__generated__/iam/ViewerMembershipDropdownSignOutMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const fragment = graphql`
  fragment ViewerMembershipDropdownFragment on Organization {
    viewer @required(action: THROW) {
      fullName
      identity @required(action: THROW) {
        email
        canListAPIKeys: permission(action: "iam:personal-api-key:list")
        canListOAuth2AccessTokens: permission(
          action: "iam:oauth2-access-token:list"
        )
      }
    }
  }
`;

const signOutMutation = graphql`
  mutation ViewerMembershipDropdownSignOutMutation {
    signOut {
      success
    }
  }
`;

export function ViewerMembershipDropdown(props: {
  fKey: ViewerMembershipDropdownFragment$key;
}) {
  const { fKey } = props;

  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { toast } = useToast();

  const {
    viewer: {
      fullName,
      identity: { canListAPIKeys, canListOAuth2AccessTokens, email },
    },
  } = useFragment<ViewerMembershipDropdownFragment$key>(fragment, fKey);
  const [signOut] = useMutation<ViewerMembershipDropdownSignOutMutation>(signOutMutation);

  const handleLogout: React.MouseEventHandler<HTMLAnchorElement> = (e) => {
    e.preventDefault();

    signOut({
      variables: {},
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: __("Request failed"),
            description: formatError(__("Cannot sign out"), e),
            variant: "error",
          });
          return;
        }
        window.location.href = "/auth/login";
      },
      onError: (e) => {
        toast({
          title: __("Error"),
          description: e.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <UserDropdown fullName={fullName} email={email}>
      {canListAPIKeys && (
        <UserDropdownItem
          to="/me/api-keys"
          icon={IconKey}
          label={__("API Keys")}
        />
      )}
      {canListOAuth2AccessTokens && (
        <UserDropdownItem
          to="/me/oauth-tokens"
          icon={IconKey}
          label={__("OAuth tokens")}
        />
      )}
      <UserDropdownItem
        to={`/organizations/${organizationId}/employee`}
        icon={IconPageTextLine}
        label={__("My Signatures")}
      />
      <UserDropdownItem
        to="mailto:support@probo.com"
        icon={IconCircleQuestionmark}
        label={__("Help")}
      />
      <DropdownSeparator />
      <UserDropdownItem
        variant="danger"
        to="/logout"
        icon={IconArrowBoxLeft}
        label="Logout"
        onClick={handleLogout}
      />
    </UserDropdown>
  );
}
