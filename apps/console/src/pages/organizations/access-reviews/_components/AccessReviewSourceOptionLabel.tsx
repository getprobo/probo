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

import { Badge } from "@probo/ui/src/v2/Badge/Badge";

import { ConnectorProviderName } from "#/pages/organizations/settings/integrations/_components/ConnectorProviderName";

import { accessReviewSourceLabel } from "../_lib/sourceLabel";

interface AccessReviewSourceOptionLabelProps {
  name: string;
  accountName?: string | null;
  provider?: string | null;
  providerName?: string | null;
}

export function AccessReviewSourceOptionLabel({
  name,
  accountName,
  provider,
  providerName,
}: AccessReviewSourceOptionLabelProps) {
  return (
    <span className="inline-flex min-w-0 items-center gap-2 text-sm">
      <span className="truncate">
        {accessReviewSourceLabel({ name, accountName })}
      </span>
      {providerName != null && (
        <Badge size={1} className="shrink-0 uppercase">
          {provider == null
            ? providerName
            : (
                <ConnectorProviderName
                  provider={provider}
                  name={providerName}
                  flagSize={12}
                />
              )}
        </Badge>
      )}
    </span>
  );
}
