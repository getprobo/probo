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
import {
  FullNameRequiredError,
  NDASignatureRequiredError,
  UnAuthenticatedError,
} from "@probo/relay";
import { Button, IconChevronLeft, IconPageCross } from "@probo/ui";
import { Link, Navigate, useLocation, useRouteError } from "react-router";

import { getPathPrefix } from "#/utils/pathPrefix";

export function DocumentPageErrorBoundary() {
  const error = useRouteError();
  const location = useLocation();
  const { __ } = useTranslate();

  const search = new URLSearchParams();

  if (location.pathname !== (getPathPrefix() || "/") || location.search !== "") {
    search.set("continue", window.location.href);
  }

  const queryString = search.toString();

  if (error instanceof UnAuthenticatedError) {
    return (
      <Navigate
        replace
        to={{
          pathname: "/connect",
          search: queryString ? "?" + queryString : "",
        }}
      />
    );
  }

  if (error instanceof FullNameRequiredError) {
    return (
      <Navigate
        replace
        to={{
          pathname: "/full-name",
          search: queryString ? "?" + queryString : "",
        }}
      />
    );
  }

  if (error instanceof NDASignatureRequiredError) {
    return (
      <Navigate
        replace
        to={{
          pathname: "/nda",
          search: queryString ? "?" + queryString : "",
        }}
      />
    );
  }

  return (
    <div className="flex flex-col h-screen bg-level-2">
      <header className="flex items-center h-12 gap-3 border-b border-border-solid px-4 flex-none bg-level-1">
        <Link
          to="/documents"
          className="size-8 grid place-items-center hover:bg-secondary-hover rounded-sm transition-all"
        >
          <IconChevronLeft size={16} />
        </Link>
      </header>
      <main className="flex-1 min-h-0 flex items-center justify-center">
        <div className="text-center max-w-sm">
          <IconPageCross size={32} className="mx-auto text-txt-tertiary mb-4" />
          <h2 className="text-lg font-medium mb-2">
            {__("Document not found")}
          </h2>
          <p className="text-sm text-txt-secondary mb-6">
            {__("The document you are looking for does not exist or has been removed.")}
          </p>
          <Button variant="secondary" asChild className="inline-flex">
            <Link to="/documents">
              {__("Back to documents")}
            </Link>
          </Button>
        </div>
      </main>
    </div>
  );
}
