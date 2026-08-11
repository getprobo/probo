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
import { Button, ErrorLayout, useToast } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { NoOrganizationAccessSignOutMutation } from "#/__generated__/iam/NoOrganizationAccessSignOutMutation.graphql";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

const signOutMutation = graphql`
  mutation NoOrganizationAccessSignOutMutation {
    signOut {
      success
    }
  }
`;

function NoOrganizationAccessContent() {
  const { t } = useTranslation();
  const { toast } = useToast();

  const [signOut, isSigningOut]
    = useMutation<NoOrganizationAccessSignOutMutation>(signOutMutation);

  const handleSignOut = () => {
    signOut({
      variables: {},
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: t("noOrganizationAccess.errors.requestFailed"),
            description: formatError(
              t("noOrganizationAccess.errors.cannotSignOut"),
              e,
            ),
            variant: "error",
          });
          return;
        }
        window.location.href = "/auth/login";
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
    <ErrorLayout
      fullPage
      showLogo
      title={t("noOrganizationAccess.title")}
      description={t("noOrganizationAccess.description")}
      actions={(
        <Button onClick={handleSignOut} disabled={isSigningOut}>
          {t("noOrganizationAccess.actions.signOut")}
        </Button>
      )}
    />
  );
}

// NoOrganizationAccess renders outside the relay providers when an error
// boundary catches a MembershipRequiredError, so it carries its own
// environment for the sign-out mutation.
export function NoOrganizationAccess() {
  return (
    <IAMRelayProvider>
      <NoOrganizationAccessContent />
    </IAMRelayProvider>
  );
}
