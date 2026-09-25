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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { useTranslation } from "react-i18next";

import type { ConnectVendorPageQuery } from "#/__generated__/core/ConnectVendorPageQuery.graphql";
import { ConnectorDocumentationLink } from "#/pages/organizations/access-reviews/dialogs/_components/ConnectorDocumentationLink";

export type ConnectVendorDriver = ConnectVendorPageQuery["response"]["accessReviewDrivers"][number];

export function ConnectFormFooter({
  documentationUrl,
  disabled,
  loading,
}: {
  documentationUrl: string | null | undefined;
  disabled?: boolean;
  loading?: boolean;
}) {
  const { t } = useTranslation("organizations/settings/integrations");

  return (
    <div className="flex items-center justify-between gap-2">
      <ConnectorDocumentationLink url={documentationUrl} variant="button" />
      <Button type="submit" variant="solid" disabled={disabled} loading={loading}>
        {t("marketplacePage.connect")}
      </Button>
    </div>
  );
}
