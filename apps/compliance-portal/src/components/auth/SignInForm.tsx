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
import { Toast } from "@base-ui/react/toast";
import type { GraphQLError } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import { getSafeContinueUrl } from "#/lib/auth/continueUrl";
import { useMutation } from "#/lib/relay/useMutation";

import type { SignInFormMutation } from "./__generated__/SignInFormMutation.graphql";
import { OIDCProviders } from "./OIDCProviders";

const RESEND_COOLDOWN_SECONDS = 60;

const sendMagicLinkMutation = graphql`
  mutation SignInFormMutation($input: SendMagicLinkInput!) {
    sendMagicLink(input: $input) {
      success
    }
  }
`;

interface SignInFormProps {
  // Absolute URL to return to after authentication (carries the request-all
  // marker so an access request resumes once signed in).
  continueTo: string;
  onCancel: () => void;
}

// Sign-in form used inside the dialog: SSO providers, then a magic-link email
// flow. On success it flips to a "check your email" state with a resend timer.
export function SignInForm({ continueTo, onCancel }: SignInFormProps) {
  const { t } = useTranslation();
  const toast = Toast.useToastManager();
  const [magicLinkSent, setMagicLinkSent] = useState(false);
  const [secondsLeft, setSecondsLeft] = useState(RESEND_COOLDOWN_SECONDS);
  const intervalRef = useRef<ReturnType<typeof setInterval>>(undefined);

  const [sendMagicLink, isSending] = useMutation<SignInFormMutation>(
    sendMagicLinkMutation,
    { errorToast: false },
  );

  useEffect(() => {
    if (!magicLinkSent) {
      return;
    }
    intervalRef.current = setInterval(() => {
      setSecondsLeft(seconds => Math.max(seconds - 1, 0));
    }, 1000);
    return () => clearInterval(intervalRef.current);
  }, [magicLinkSent]);

  const handleSend = (email: string) => {
    void sendMagicLink({
      variables: { input: { email, continue: continueTo } },
      onCompleted: (_response, errors) => {
        const code = (errors?.[0] as GraphQLError | undefined)?.extensions?.code;

        // Already signed in elsewhere: jump straight to the return URL so any
        // pending access request resumes.
        if (code === "ALREADY_AUTHENTICATED") {
          window.location.href = getSafeContinueUrl(continueTo);
          return;
        }

        if (errors && errors.length > 0) {
          toast.add({ title: t("auth.errors.magicLinkFailed"), type: "error" });
          return;
        }

        setSecondsLeft(RESEND_COOLDOWN_SECONDS);
        setMagicLinkSent(true);
        toast.add({ title: t("auth.signIn.magicLinkSent"), type: "success" });
      },
      onError: () => {
        toast.add({ title: t("auth.errors.magicLinkFailed"), type: "error" });
      },
    }).catch(() => {});
  };

  const resendDisabled = magicLinkSent && secondsLeft > 0;
  const submitLabel = magicLinkSent
    ? secondsLeft > 0
      ? t("auth.signIn.resendIn", { seconds: secondsLeft })
      : t("auth.signIn.resend")
    : t("auth.signIn.sendMagicLink");

  return (
    <Form
      className="flex flex-col gap-4"
      onFormSubmit={(values) => {
        handleSend(String(values.email ?? ""));
      }}
    >
      <DialogBody className="flex flex-col gap-6">
        <OIDCProviders continueTo={continueTo} />

        <Field.Root name="email" className="flex flex-col gap-1.5">
          <Field.Label className="text-1 font-medium text-sand-12">
            {t("auth.signIn.emailLabel")}
          </Field.Label>
          <TextField
            type="email"
            name="email"
            required
            placeholder={t("auth.signIn.emailPlaceholder")}
          />
          <Field.Error className="text-1 text-red-11" match="valueMissing">
            {t("auth.signIn.emailRequired")}
          </Field.Error>
          <Field.Error className="text-1 text-red-11" match="typeMismatch">
            {t("auth.signIn.emailInvalid")}
          </Field.Error>
        </Field.Root>

        {magicLinkSent && (
          <Text size={1} color="neutral">
            {t("auth.signIn.magicLinkSentNote")}
          </Text>
        )}
      </DialogBody>

      <DialogFooter>
        <Button type="button" variant="soft" color="neutral" highContrast onClick={onCancel}>
          {t("common.cancel")}
        </Button>
        <Button
          type="submit"
          variant="solid"
          color="neutral"
          highContrast
          loading={isSending}
          disabled={resendDisabled}
        >
          {submitLabel}
        </Button>
      </DialogFooter>
    </Form>
  );
}
