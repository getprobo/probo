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

import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import { useMutation } from "#/lib/relay/useMutation";

import type { useSignOutMutation } from "./__generated__/useSignOutMutation.graphql";

const signOutMutation = graphql`
  mutation useSignOutMutation {
    signOut {
      success
    }
  }
`;

// Closes the trust-center session and clears the session cookie. Callers reload
// the page after success so the UI drops back to the guest chrome.
export function useSignOut() {
  const { t } = useTranslation();
  const [commit, isSigningOut] = useMutation<useSignOutMutation>(signOutMutation, {
    errorToast: t("userMenu.signOutFailed"),
  });

  const signOut = useCallback(async () => {
    try {
      await commit({ variables: {} });
      window.location.reload();
    } catch {
      // errorToast already handles user-facing feedback.
    }
  }, [commit]);

  return [signOut, isSigningOut] as const;
}
