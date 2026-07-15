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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Field, Google, Microsoft, useToast } from "@probo/ui";
import { useEffect, useRef, useState } from "react";
import { useSearchParams } from "react-router";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const schema = z.object({
  email: z.string().email(),
});

type FormData = z.infer<typeof schema>;

const timerDurationSeconds = 60;

type OIDCProvider = {
  name: string;
  loginURL: string;
};

function buildAuthorizeContinueURL(authorizeParam: string | null): string | null {
  if (!authorizeParam) {
    return null;
  }

  const url = new URL("/api/connect/v1/oauth2/authorize", window.location.origin);
  const params = new URLSearchParams(authorizeParam);
  for (const [key, value] of params.entries()) {
    url.searchParams.set(key, value);
  }

  return url.toString();
}

async function fetchOIDCProviders(): Promise<OIDCProvider[]> {
  const response = await fetch("/api/connect/v1/graphql", {
    method: "POST",
    headers: { "content-type": "application/json" },
    credentials: "include",
    body: JSON.stringify({
      query: "query { oidcProviders { name loginURL } }",
    }),
  });

  if (!response.ok) {
    return [];
  }

  const payload = await response.json() as {
    data?: { oidcProviders?: OIDCProvider[] };
  };

  return payload.data?.oidcProviders ?? [];
}

export default function PortalLoginPage() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();
  const authorizeParam = searchParams.get("authorize");
  const authorizeContinueURL = buildAuthorizeContinueURL(authorizeParam);

  const [magicLinkSent, setMagicLinkSent] = useState(false);
  const interval = useRef<ReturnType<typeof setTimeout>>(undefined);
  const [timer, setTimer] = useState(timerDurationSeconds);
  const [oidcProviders, setOidcProviders] = useState<OIDCProvider[]>([]);

  usePageTitle(__("Sign in to Compliance Page"));

  useEffect(() => {
    void fetchOIDCProviders().then(setOidcProviders);
  }, []);

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
    if (!authorizeParam) {
      toast({
        title: __("Error"),
        description: __("Invalid sign-in request"),
        variant: "error",
      });
      return;
    }

    const body = new URLSearchParams();
    body.set("email", email);
    body.set("authorize", authorizeParam);

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

  const providerIcons: Record<string, typeof Google> = {
    google: Google,
    microsoft: Microsoft,
  };

  if (!authorizeContinueURL) {
    return (
      <p className="text-txt-tertiary text-center">
        {__("Invalid sign-in request")}
      </p>
    );
  }

  return (
    <div className="space-y-6 w-full">
      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-bold">{__("Sign in to Compliance Page")}</h1>
        <p className="text-txt-tertiary">
          {__("Use your email or a connected account to continue")}
        </p>
      </div>

      {oidcProviders.length > 0 && (
        <div className="space-y-3">
          {oidcProviders.map(provider => {
            const Icon = providerIcons[provider.name];
            const loginURL = new URL(provider.loginURL, window.location.origin);
            loginURL.searchParams.set("continue", authorizeContinueURL);

            return (
              <Button
                key={provider.name}
                variant="secondary"
                className="w-full h-10"
                onClick={() => {
                  window.location.href = loginURL.toString();
                }}
              >
                <span className="flex items-center gap-2">
                  {Icon && <Icon width={18} height={18} />}
                  {__(`Sign in with ${provider.name.charAt(0).toUpperCase() + provider.name.slice(1)}`)}
                </span>
              </Button>
            );
          })}
        </div>
      )}

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
    </div>
  );
}
