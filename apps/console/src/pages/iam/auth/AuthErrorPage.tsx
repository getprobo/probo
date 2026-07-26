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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button } from "@probo/ui";
import { useSearchParams } from "react-router";

type AuthErrorContent = {
  title: string;
  description: string;
};

function useAuthErrorContent(code: string | null): AuthErrorContent {
  const { __ } = useTranslate();

  switch (code) {
    case "personal_account_not_allowed":
      return {
        title: __("Enterprise account required"),
        description: __(
          "Personal Google and Microsoft accounts cannot be used to sign in. Please use your work or school account instead.",
        ),
      };
    case "email_not_verified":
      return {
        title: __("Email not verified"),
        description: __(
          "Your email address is not verified with the identity provider. Please verify it, then try signing in again.",
        ),
      };
    case "invalid_state":
      return {
        title: __("Sign-in session expired"),
        description: __(
          "This sign-in attempt is no longer valid. Please start again from the sign-in page.",
        ),
      };
    case "magic_link_expired":
      return {
        title: __("Link Expired"),
        description: __(
          "This magic link has expired. Magic links are only valid for 15 minutes. Please request a new one.",
        ),
      };
    case "magic_link_already_used":
      return {
        title: __("Link Already Used"),
        description: __(
          "This magic link has already been used. Please request a new one.",
        ),
      };
    case "magic_link_invalid":
      return {
        title: __("Invalid link"),
        description: __(
          "This magic link is invalid. Please request a new one.",
        ),
      };
    default:
      return {
        title: __("Authentication failed"),
        description: __(
          "We could not complete your sign-in. Please try again.",
        ),
      };
  }
}

export default function AuthErrorPage() {
  const { __ } = useTranslate();
  const [searchParams] = useSearchParams();
  const content = useAuthErrorContent(searchParams.get("error"));

  const continueParam = searchParams.get("continue");
  const loginSearch = continueParam
    ? `?${new URLSearchParams({ continue: continueParam }).toString()}`
    : "";

  usePageTitle(content.title);

  return (
    <div className="space-y-6 w-full">
      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-bold">{content.title}</h1>
        <p className="text-txt-tertiary">{content.description}</p>
      </div>
      <Button
        className="w-full h-10"
        to={{
          pathname: "/auth/login",
          search: loginSearch,
        }}
      >
        {__("Sign in")}
      </Button>
    </div>
  );
}
