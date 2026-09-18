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
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router";

export default function MagicLinkPage() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const token = searchParams.get("token")?.trim() ?? "";

  usePageTitle(t("magicLinkPage.pageTitle"));

  async function handleContinue(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);

    const showContinueError = () => {
      toast({
        title: t("common.error"),
        description: t("magicLinkPage.errors.continue"),
        variant: "error",
      });
    };

    try {
      const body = new URLSearchParams();
      body.set("token", token);

      const url = new URL("/api/connect/v1/magic-link/verify", window.location.origin);
      let response: Response;
      try {
        response = await fetch(url, {
          method: "POST",
          headers: {
            "accept": "application/json",
            "content-type": "application/x-www-form-urlencoded",
          },
          credentials: "include",
          body,
        });
      } catch {
        showContinueError();
        return;
      }

      if (!response.ok) {
        showContinueError();
        return;
      }

      let redirectURL = "";
      try {
        const payload: unknown = await response.json();
        if (
          payload !== null
          && typeof payload === "object"
          && "redirect_url" in payload
          && typeof payload.redirect_url === "string"
        ) {
          redirectURL = payload.redirect_url;
        }
      } catch {
        showContinueError();
        return;
      }

      if (redirectURL === "") {
        showContinueError();
        return;
      }

      window.location.assign(new URL(redirectURL, window.location.origin).href);
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="flex w-full flex-col gap-8">
      <div className="flex flex-col gap-1">
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {t("magicLinkPage.title")}
        </Heading>
        <Text size={2} align="center" className="block">
          {t("magicLinkPage.description")}
        </Text>
      </div>

      {token === ""
        ? (
            <ButtonLink
              to="/auth/login"
              variant="solid"
              color="neutral"
              highContrast
              size={3}
              className="w-full"
            >
              {t("auth.actions.signIn")}
            </ButtonLink>
          )
        : (
            <form
              className="flex flex-col"
              onSubmit={(event) => {
                void handleContinue(event);
              }}
            >
              <Button
                type="submit"
                variant="solid"
                color="neutral"
                highContrast
                size={3}
                className="w-full"
                loading={isSubmitting}
              >
                {t("magicLinkPage.actions.continue")}
              </Button>
            </form>
          )}
    </div>
  );
}
