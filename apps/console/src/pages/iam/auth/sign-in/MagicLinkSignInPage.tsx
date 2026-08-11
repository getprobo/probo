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

import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Field, IconChevronLeft, useToast } from "@probo/ui";
import { type FormEventHandler, useState } from "react";
import { useMutation } from "react-relay";
import { Link, useLocation, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import type { MagicLinkSignInPageMutation } from "#/__generated__/iam/MagicLinkSignInPageMutation.graphql";
import { useSafeContinueUrl } from "#/hooks/useSafeContinueUrl";

const sendMagicLinkMutation = graphql`
  mutation MagicLinkSignInPageMutation($input: SendMagicLinkInput!) {
    sendMagicLink(input: $input) {
      success
    }
  }
`;

export default function MagicLinkSignInPage() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const safeContinueUrl = useSafeContinueUrl();

  const [sending, setSending] = useState(false);
  const [sentTo, setSentTo] = useState<string | null>(null);

  usePageTitle(__("Sign in with a magic link"));

  const [sendMagicLink]
    = useMutation<MagicLinkSignInPageMutation>(sendMagicLinkMutation);

  const handleSubmit: FormEventHandler<HTMLFormElement> = (e) => {
    e.preventDefault();

    const formData = new FormData(e.currentTarget);
    const email = (formData.get("email") as string | null)?.trim() ?? "";

    if (!email) return;

    setSending(true);

    sendMagicLink({
      variables: {
        input: {
          email,
          organizationId: searchParams.get("organization-id"),
          continue: safeContinueUrl.toString(),
        },
      },
      onCompleted: (_, errors) => {
        setSending(false);

        if (errors) {
          toast({
            title: __("Error"),
            description: formatError(__("Failed to send magic link"), errors),
            variant: "error",
          });
          return;
        }

        setSentTo(email);
      },
      onError: (err) => {
        setSending(false);
        toast({
          title: __("Error"),
          description: err.message || __("Failed to send magic link"),
          variant: "error",
        });
      },
    });
  };

  // The server answers identically whether or not the address has an account,
  // so this confirmation is deliberately vague about whether one was sent.
  if (sentTo) {
    return (
      <div className="space-y-6 w-full max-w-md mx-auto pt-8 text-center">
        <h1 className="text-2xl font-bold">{__("Check your email")}</h1>
        <p className="text-txt-tertiary">
          {__("If an account exists for")}
          {" "}
          <span className="text-txt-primary">{sentTo}</span>
          {", "}
          {__(
            "we've sent it a sign-in link. The link expires in 15 minutes and can only be used once.",
          )}
        </p>
        <Button
          variant="secondary"
          className="w-xs h-10 mx-auto"
          onClick={() => setSentTo(null)}
        >
          {__("Use a different email")}
        </Button>
      </div>
    );
  }

  return (
    <form
      className="space-y-6 w-full max-w-md mx-auto pt-4"
      onSubmit={handleSubmit}
    >
      <Link
        to={{ pathname: "/auth/login", search: location.search }}
        className="flex items-center gap-2 text-txt-secondary hover:text-txt-primary transition-colors mb-4"
      >
        <IconChevronLeft size={20} />
        <span className="text-sm">{__("Back")}</span>
      </Link>

      <h1 className="text-center text-2xl font-bold">
        {__("Sign in with a magic link")}
      </h1>
      <p className="text-center text-txt-tertiary mt-1 mb-6">
        {__("We'll email you a link that signs you in — no password needed")}
      </p>

      <Field
        required
        placeholder={__("Work Email")}
        name="email"
        type="email"
        label={__("Work Email")}
        autoFocus
      />

      <Button className="w-xs h-10 mx-auto mt-6" disabled={sending}>
        {sending ? __("Sending...") : __("Send magic link")}
      </Button>
    </form>
  );
}
