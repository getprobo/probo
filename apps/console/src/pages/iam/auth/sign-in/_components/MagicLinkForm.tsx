// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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

    const response = await fetch("/api/connect/v1/magic-link/send", {
      method: "POST",
      headers: { "content-type": "application/x-www-form-urlencoded" },
      credentials: "include",
      body,
    });

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
