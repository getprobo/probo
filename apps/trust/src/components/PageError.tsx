// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
import {
  Button,
  ErrorDetailMessage,
  ErrorDetails,
  ErrorLayout,
} from "@probo/ui";
import { useEffect, useRef } from "react";
import { Link, useLocation, useRouteError } from "react-router";

type Props = {
  resetErrorBoundary?: () => void;
  error?: Error;
};

export function PageError({ resetErrorBoundary, error: propsError }: Props) {
  const routeError = useRouteError();
  const error = routeError ?? propsError;
  const { __ } = useTranslate();
  const location = useLocation();
  const baseLocation = useRef(location);

  const isFullPage = Boolean(routeError ?? propsError);

  // Reset error boundary on page change
  useEffect(() => {
    if (
      location.pathname !== baseLocation.current.pathname
      && resetErrorBoundary
    ) {
      resetErrorBoundary();
    }
  }, [location, resetErrorBoundary]);

  const actions = (
    <Button asChild>
      <Link to="/">{__("Go home")}</Link>
    </Button>
  );

  const layoutProps = {
    fullPage: isFullPage,
    showLogo: isFullPage,
    actions,
  };

  if (!error) {
    return (
      <ErrorLayout
        {...layoutProps}
        title={__("Page not found")}
        description={__("The page you are looking for does not exist.")}
      />
    );
  }

  if (error instanceof Error
    && error.message
      .toLowerCase()
      .match(/(token|expired|invalid|401|unauthorized)/)
  ) {
    const isExpiredToken = error.message.toLowerCase().includes("expired");
    const title = isExpiredToken
      ? __("Expired token")
      : __("Invalid Access Link");
    const description = isExpiredToken
      ? __(
          "This access link has expired. Compliance page access links are valid for 7 days for security reasons.",
        )
      : __(
          "This access link is not valid. It may have been revoked or the link might be incorrect.",
        );
    return (
      <ErrorLayout
        {...layoutProps}
        title={title}
        description={description}
      />
    );
  }

  return (
    <ErrorLayout
      {...layoutProps}
      title={__("Something went wrong")}
      description={__("We hit an unexpected error. Head back home to continue.")}
    >
      {error instanceof Error && (
        <ErrorDetails summary={__("Technical details")}>
          <ErrorDetailMessage>{error.message}</ErrorDetailMessage>
        </ErrorDetails>
      )}
    </ErrorLayout>
  );
}
