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

import { useTranslate } from "@probo/i18n";
import { Button, Field, useToast } from "@probo/ui";
import { useEffect, useRef, useState } from "react";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { usePostAuthRedirectUrl } from "#/hooks/usePostAuthRedirectUrl";

const schema = z.object({
  email: z.string().email(),
});

type FormData = z.infer<typeof schema>;

const timerDurationSeconds = 60;

export function MagicLinkForm() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const postAuthRedirectUrl = usePostAuthRedirectUrl();

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

  const {
    handleSubmit: handleSubmitWrapper,
    register,
    formState,
  } = useFormWithSchema(schema, {
    defaultValues: { email: "" },
  });

  const handleSubmit = handleSubmitWrapper(async ({ email }: FormData) => {
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
        title: __("Error"),
        description: __("Cannot send magic link"),
        variant: "error",
      });
      return;
    }

    if (!response.ok) {
      toast({
        title: __("Error"),
        description: __("Cannot send magic link"),
        variant: "error",
      });
      return;
    }

    toast({
      title: __("Success"),
      description: __("Magic link sent!"),
      variant: "success",
    });
    setTimer(timerDurationSeconds);
    setMagicLinkSent(true);
  });

  return (
    <form onSubmit={e => void handleSubmit(e)} className="space-y-4">
      <Field
        label={__("Email")}
        placeholder="john.doe@acme.com"
        {...register("email")}
        type="email"
        required
        error={formState.errors.email?.message}
      />

      {magicLinkSent && (
        <p className="text-txt-primary text-sm">
          {__(
            "Magic link sent! Check your email and use the link to continue.",
          )}
        </p>
      )}

      <Button
        type="submit"
        className="w-full h-10"
        disabled={formState.isSubmitting || (magicLinkSent && timer !== 0)}
      >
        {magicLinkSent
          ? timer === 0
            ? __("Resend Link")
            : `${__("Resend Link in")} ${timer}s`
          : __("Send Magic Link")}
      </Button>
    </form>
  );
}
