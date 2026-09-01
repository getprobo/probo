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

import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import type { ReactNode } from "react";
import { Suspense, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useQueryLoader } from "react-relay";
import { useParams, useSearchParams } from "react-router";

import type { BindPageQuery } from "#/__generated__/core/BindPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { PageHeader } from "#/pages/_components/PageHeader";

import { BindTokenError } from "./_components/BindTokenError";
import { bindingsPage } from "./_components/variants";
import { BindPage, bindPageQuery } from "./BindPage";
import { BindPageSkeleton } from "./BindPageSkeleton";

interface BindPageQueryLoaderProps {
  token: string;
}

function BindPageQueryLoader({ token }: BindPageQueryLoaderProps) {
  const [queryRef, loadQuery] = useQueryLoader<BindPageQuery>(bindPageQuery);

  useEffect(() => {
    loadQuery({ token }, { fetchPolicy: "network-only" });
  }, [loadQuery, token]);

  if (queryRef === undefined || queryRef === null) {
    return <BindPageSkeleton />;
  }

  return <BindPage queryRef={queryRef} token={token} />;
}

interface BindPageShellProps {
  children: ReactNode;
}

function BindPageShell({ children }: BindPageShellProps) {
  const { t } = useTranslation();
  const { t: tBindings } = useTranslation("bindings");
  const { organizationId } = useParams();
  const slots = bindingsPage();

  if (organizationId === undefined) {
    throw new NotFoundError("organizationId is required");
  }

  return (
    <main className={slots.main()}>
      <PageHeader
        homeLabel={t("homePage.breadcrumb")}
        parent={{
          label: tBindings("list.breadcrumb"),
          to: `/${organizationId}/bindings`,
        }}
        currentLabel={tBindings("bind.breadcrumb")}
        title={tBindings("bind.title")}
      />
      {children}
    </main>
  );
}

export default function BindPageLoader() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token")?.trim() ?? "";

  if (token === "") {
    return (
      <BindPageShell>
        <BindTokenError reason="missing" />
      </BindPageShell>
    );
  }

  return (
    <ErrorBoundary
      key={token}
      fallback={(
        <BindPageShell>
          <BindTokenError reason="invalid" />
        </BindPageShell>
      )}
    >
      <Suspense fallback={<BindPageSkeleton />}>
        <BindPageQueryLoader token={token} />
      </Suspense>
    </ErrorBoundary>
  );
}
