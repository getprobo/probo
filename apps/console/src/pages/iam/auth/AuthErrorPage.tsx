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
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
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

  const continueParam = searchParams.get("continue");
  const loginSearch = continueParam
    ? `?${new URLSearchParams({ continue: continueParam }).toString()}`
    : "";

  usePageTitle(content.title);

  return (
    <div className="flex w-full flex-col gap-8">
      <div className="flex flex-col gap-1">
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {content.title}
        </Heading>
        <Text size={2} align="center" className="block">
          {content.description}
        </Text>
      </div>
      <ButtonLink
        to={{
          pathname: "/auth/login",
          search: loginSearch,
        }}
        variant="solid"
        color="neutral"
        highContrast
        size={3}
        className="w-full"
      >
        {t("auth.actions.signIn")}
      </ButtonLink>
    </div>
  );
}
