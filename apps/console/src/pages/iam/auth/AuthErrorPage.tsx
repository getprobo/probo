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
import { useTranslation } from "react-i18next";
import { Button } from "@probo/ui";
import { useSearchParams } from "react-router";

type AuthErrorContent = {
  title: string;
  description: string;
};

function useAuthErrorContent(code: string | null): AuthErrorContent {
  const { t } = useTranslation();

  switch (code) {
    case "personal_account_not_allowed":
      return {
        title: t("authError.enterpriseAccountRequired.title"),
        description: t("authError.enterpriseAccountRequired.description"),
      };
    case "email_not_verified":
      return {
        title: t("authError.emailNotVerified.title"),
        description: t("authError.emailNotVerified.description"),
      };
    case "invalid_state":
      return {
        title: t("authError.signInSessionExpired.title"),
        description: t("authError.signInSessionExpired.description"),
      };
    case "magic_link_expired":
      return {
        title: t("authError.magicLinkExpired.title"),
        description: t("authError.magicLinkExpired.description"),
      };
    case "magic_link_already_used":
      return {
        title: t("authError.magicLinkAlreadyUsed.title"),
        description: t("authError.magicLinkAlreadyUsed.description"),
      };
    case "magic_link_invalid":
      return {
        title: t("authError.invalidLink.title"),
        description: t("authError.invalidLink.description"),
      };
    default:
      return {
        title: t("authError.default.title"),
        description: t("authError.default.description"),
      };
  }
}

export default function AuthErrorPage() {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const content = useAuthErrorContent(searchParams.get("error"));

  usePageTitle(content.title);

  return (
    <div className="space-y-6 w-full">
      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-bold">{content.title}</h1>
        <p className="text-txt-tertiary">{content.description}</p>
      </div>
      <Button className="w-full h-10" to="/auth/login">
        {t("auth.actions.signIn")}
      </Button>
    </div>
  );
}
