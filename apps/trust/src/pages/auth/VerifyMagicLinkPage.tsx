// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { formatError, type GraphQLError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { useToast } from "@probo/ui";
import { useCallback, useEffect, useRef } from "react";
import { useMutation } from "react-relay";
import { Link, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import { getPathPrefix } from "#/utils/pathPrefix";

import type { VerifyMagicLinkPageMutation } from "./__generated__/VerifyMagicLinkPageMutation.graphql";

const verifyMagicLinkMutation = graphql`
  mutation VerifyMagicLinkPageMutation($input: VerifyMagicLinkInput!) {
    verifyMagicLink(input: $input) {
      continue
    }
  }
`;

export default function VerifyMagicLinkPagePageMutation() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [searchParams] = useSearchParams();
  const submittedRef = useRef<boolean>(false);

  usePageTitle(__("Verify Magic Link"));

  const [verifyMagicLink] = useMutation<VerifyMagicLinkPageMutation>(
    verifyMagicLinkMutation,
  );

  const handleVerifyMagicToken = useCallback((token: string) => {
    if (submittedRef.current) return;

    verifyMagicLink({
      variables: {
        input: { token },
      },
      onCompleted: (response, errors: GraphQLError[] | null) => {
        if (errors) {
          for (const err of errors) {
            if (err.extensions?.code === "ALREADY_AUTHENTICATED") {
              window.location.href = window.location.origin + getPathPrefix();
              return;
            }
          }

          const hasExpiredToken = errors.some(
            err => err.message === "token has expired",
          );

          toast({
            title: __("Error"),
            description: hasExpiredToken
              ? __("This magic link has expired. Please request a new one.")
              : formatError(__("Failed to connect"), errors),
            variant: "error",
          });
          return;
        }

        const { verifyMagicLink } = response;

        toast({
          title: __("Success"),
          description: __("Your have successfully signed in"),
          variant: "success",
        });

        if (verifyMagicLink?.continue) {
          try {
            const continueUrl = new URL(verifyMagicLink.continue, window.location.origin);
            window.location.href = window.location.origin + continueUrl.pathname + continueUrl.search;
          } catch {
            window.location.href = window.location.origin + getPathPrefix();
          }
        } else {
          window.location.href = window.location.origin + getPathPrefix();
        }
      },
      onError: (err) => {
        toast({
          title: __("Error"),
          description: err.message,
          variant: "error",
        });
      },
    });
  }, [__, toast, verifyMagicLink]);

  useEffect(() => {
    const token = searchParams.get("token");
    if (!submittedRef.current && token) {
      void handleVerifyMagicToken(token.trim());
      submittedRef.current = true;
    }
  }, [handleVerifyMagicToken, searchParams]);

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Email Confirmation")}</h1>
        <p className="text-txt-tertiary">
          {__("Confirming your email address to complete registration…")}
        </p>
      </div>
      <div className="text-center mt-6 text-sm text-txt-secondary">
        <Link
          to="/connect"
          className="underline hover:text-txt-primary"
        >
          {__("Go back")}
        </Link>
      </div>
    </div>
  );
}
