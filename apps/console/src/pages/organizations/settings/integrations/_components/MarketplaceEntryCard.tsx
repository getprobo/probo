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

import { ThirdPartyLogo } from "@probo/ui";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { marketplacePath } from "../_lib/integrationPath";

const logoIntervalMs = 5000;

interface MarketplaceEntryCardProps {
  organizationId: string;
  providers: readonly string[];
}

export function MarketplaceEntryCard({
  organizationId,
  providers,
}: MarketplaceEntryCardProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const label = t("listPage.addMore");
  const [index, setIndex] = useState(0);
  const provider = providers[index % Math.max(providers.length, 1)];

  useEffect(() => {
    if (providers.length < 2) {
      return;
    }

    const id = window.setInterval(() => {
      setIndex(current => (current + 1) % providers.length);
    }, logoIntervalMs);

    return () => window.clearInterval(id);
  }, [providers.length]);

  return (
    <Card variant="soft" size={2} interactive className="relative h-full min-h-40 overflow-hidden">
      <Link
        to={marketplacePath(organizationId)}
        underline={false}
        className="absolute inset-0 z-10"
        aria-label={label}
      />
      <div className="flex h-full min-h-36 flex-col items-center justify-center gap-3 opacity-60">
        {provider != null && (
          <div aria-hidden className="flex size-8 items-center justify-center opacity-45">
            <ThirdPartyLogo
              key={provider}
              thirdParty={provider}
              tint
              className="size-8 motion-safe:animate-[marketplace-logo_400ms_ease-out]"
            />
            <style>
              {`@keyframes marketplace-logo {
                from { opacity: 0; }
                to { opacity: 1; }
              }`}
            </style>
          </div>
        )}
        <Text size={2} color="faint">
          {label}
        </Text>
      </div>
    </Card>
  );
}
