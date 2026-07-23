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

import {
  Button,
  ErrorDetailMessage,
  ErrorDetails,
  ErrorLayout,
} from "@probo/ui";
import { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { Link, useLocation, useRouteError } from "react-router";

type Props = {
  resetErrorBoundary?: () => void;
  error?: Error;
};

export function PageError({ resetErrorBoundary, error: propsError }: Props) {
  const routeError = useRouteError();
  const error = routeError ?? propsError;
  const { t } = useTranslation();
  const location = useLocation();
  const baseLocation = useRef(location);

  const isFullPage = Boolean(routeError ?? propsError);
  const isEmbeddedNotFound = !isFullPage
    && /^\/organizations\/[^/]+/.test(location.pathname);

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
      <Link to="/">{t("pageError.actions.goHome")}</Link>
    </Button>
  );

  const layoutProps = {
    fullPage: isFullPage || !isEmbeddedNotFound,
    showLogo: isFullPage,
    actions,
  };

  if (!error || (error instanceof Error && error.message.includes("PAGE_NOT_FOUND"))) {
    return (
      <ErrorLayout
        {...layoutProps}
        title={t("pageError.notFound.title")}
        description={t("pageError.notFound.description")}
      />
    );
  }

  if (error instanceof Error && error.message.includes("FORBIDDEN")) {
    return (
      <ErrorLayout
        {...layoutProps}
        title={t("pageError.notFound.title")}
        description={t("pageError.notFound.description")}
      />
    );
  }

  return (
    <ErrorLayout
      {...layoutProps}
      title={t("pageError.unexpected.title")}
      description={t("pageError.unexpected.description")}
    >
      {error instanceof Error && (
        <ErrorDetails summary={t("pageError.technicalDetails")}>
          <ErrorDetailMessage>{error.message}</ErrorDetailMessage>
        </ErrorDetails>
      )}
    </ErrorLayout>
  );
}
