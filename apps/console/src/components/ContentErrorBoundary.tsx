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

import { ForbiddenError } from "@probo/relay";
import { ErrorDetailMessage, ErrorDetails, ErrorLayout } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useRouteError } from "react-router";

// Route-level boundary for a child route rendered inside a layout: it replaces
// only the outlet content, so the surrounding header and tabs stay usable when
// a member opens a page their role cannot reach.
export function ContentErrorBoundary() {
  const error = useRouteError();
  const { t } = useTranslation();

  if (error instanceof ForbiddenError) {
    return (
      <ErrorLayout
        title={t("pageError.forbidden.title")}
        description={t("pageError.forbidden.description")}
      />
    );
  }

  return (
    <ErrorLayout
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
