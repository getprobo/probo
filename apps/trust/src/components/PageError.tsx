// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
