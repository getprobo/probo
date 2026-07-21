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
import {
  DropdownSeparator,
  IconArrowBoxLeft,
  IconCircleQuestionmark,
  IconKey,
  UserDropdown,
  UserDropdownItem,
  useToast,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { ViewerDropdownFragment$key } from "#/__generated__/iam/ViewerDropdownFragment.graphql";
import type { ViewerDropdownSignOutMutation } from "#/__generated__/iam/ViewerDropdownSignOutMutation.graphql";

export const fragment = graphql`
  fragment ViewerDropdownFragment on Identity {
    canListAPIKeys: permission(action: "iam:personal-api-key:list")
    canListOAuth2AccessTokens: permission(action: "iam:oauth2-access-token:list")
    email
    fullName
  }
`;

const signOutMutation = graphql`
  mutation ViewerDropdownSignOutMutation {
    signOut {
      success
    }
  }
`;

export function ViewerDropdown(props: { fKey: ViewerDropdownFragment$key }) {
  const { fKey } = props;

  const { t } = useTranslation();
  const { toast } = useToast();

  const { canListAPIKeys, canListOAuth2AccessTokens, email, fullName }
    = useFragment<ViewerDropdownFragment$key>(fragment, fKey);
  const [signOut] = useMutation<ViewerDropdownSignOutMutation>(signOutMutation);

  const handleLogout: React.MouseEventHandler<HTMLAnchorElement> = (e) => {
    e.preventDefault();

    signOut({
      variables: {},
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: t("viewerDropdown.errors.requestFailed"),
            description: formatError(t("viewerDropdown.errors.cannotSignOut"), e),
            variant: "error",
          });
          return;
        }
        window.location.reload();
      },
      onError: (e) => {
        toast({
          title: t("common.error"),
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
          label={t("apiKeys.title")}
        />
      )}
      {canListOAuth2AccessTokens && (
        <UserDropdownItem
          to="/me/oauth-tokens"
          icon={IconKey}
          label={t("viewerDropdown.actions.oauthTokens")}
        />
      )}
      <UserDropdownItem
        to="mailto:support@probo.com"
        icon={IconCircleQuestionmark}
        label={t("viewerDropdown.actions.help")}
      />
      <DropdownSeparator />
      <UserDropdownItem
        variant="danger"
        to="/logout"
        icon={IconArrowBoxLeft}
        label={t("viewerDropdown.actions.logout")}
        onClick={handleLogout}
      />
    </UserDropdown>
  );
}
