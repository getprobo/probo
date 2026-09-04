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

import { Field } from "@base-ui/react/field";
import { Form } from "@base-ui/react/form";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchQuery, graphql, useRelayEnvironment } from "react-relay";

import type { MagicLinkFormSSOQuery } from "#/__generated__/iam/MagicLinkFormSSOQuery.graphql";
import { usePostAuthRedirectUrl } from "#/hooks/usePostAuthRedirectUrl";

const magicLinkFormSSOQuery = graphql`
  query MagicLinkFormSSOQuery($email: EmailAddr!) {
    ssoLoginURL(email: $email) @catch(to: RESULT)
  }
`;

const timerDurationSeconds = 60;

export function MagicLinkForm() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const environment = useRelayEnvironment();
  const postAuthRedirectUrl = usePostAuthRedirectUrl();

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [magicLinkSent, setMagicLinkSent] = useState(false);
  const interval = useRef<ReturnType<typeof setInterval>>(undefined);
  const [timer, setTimer] = useState(timerDurationSeconds);

  useEffect(() => {
    if (!magicLinkSent && interval.current) {
      clearInterval(interval.current);
      interval.current = undefined;
    }
    if (magicLinkSent) {
      clearInterval(interval.current);
      interval.current = setInterval(() => {
        setTimer(value => Math.max(value - 1, 0));
      }, 1000);
    }

    return () => {
      clearInterval(interval.current);
    };
  }, [magicLinkSent]);

  const handleSubmit = async (email: string) => {
    setIsSubmitting(true);
    try {
      try {
        const data = await fetchQuery<MagicLinkFormSSOQuery>(
          environment,
          magicLinkFormSSOQuery,
          { email },
          { fetchPolicy: "network-only" },
        ).toPromise();

        const ssoLoginURL = data?.ssoLoginURL;
        if (ssoLoginURL?.ok && ssoLoginURL.value) {
          const url = new URL(ssoLoginURL.value, window.location.origin);
          url.searchParams.set("continue", postAuthRedirectUrl);
          window.location.href = url.toString();
          return;
        }
      } catch {
        // No unique SAML config (or the lookup failed): send a magic link.
      }

      const body = new URLSearchParams();
      body.set("email", email);
      body.set("continue", postAuthRedirectUrl);

      let response: Response;
      try {
        response = await fetch("/api/connect/v1/magic-link/send", {
          method: "POST",
          headers: { "content-type": "application/x-www-form-urlencoded" },
          credentials: "include",
          body,
        });
      } catch {
        toast({
          title: t("common.error"),
          description: t("magicLinkForm.errors.send"),
          variant: "error",
        });
        return;
      }

      if (!response.ok) {
        toast({
          title: t("common.error"),
          description: t("magicLinkForm.errors.send"),
          variant: "error",
        });
        return;
      }

      toast({
        title: t("common.success"),
        description: t("magicLinkForm.messages.sent"),
        variant: "success",
      });
      setTimer(timerDurationSeconds);
      setMagicLinkSent(true);
    } finally {
      setIsSubmitting(false);
    }
  };

  const waitingToResend = magicLinkSent && timer !== 0;

  return (
    <Form
      className="flex flex-col gap-5"
      onFormSubmit={(values) => {
        void handleSubmit(String(values.email ?? ""));
      }}
    >
      <Field.Root name="email" className="flex flex-col gap-1.5">
        <Field.Label className="text-1 font-medium text-sand-12">
          {t("magicLinkForm.fields.email")}
        </Field.Label>
        <TextField
          type="email"
          name="email"
          required
          placeholder="john.doe@acme.com"
        />
        <Field.Error className="text-1 text-red-11" />
      </Field.Root>

      {magicLinkSent && (
        <Text size={2} className="block">
          {t("magicLinkForm.sentDescription")}
        </Text>
      )}

      <Button
        type="submit"
        variant="solid"
        color="neutral"
        highContrast
        size={3}
        className="w-full"
        loading={isSubmitting}
        disabled={waitingToResend}
      >
        {magicLinkSent
          ? waitingToResend
            ? t("magicLinkForm.actions.resendIn", { count: timer })
            : t("magicLinkForm.actions.resend")
          : t("magicLinkForm.actions.send")}
      </Button>
    </Form>
  );
}
