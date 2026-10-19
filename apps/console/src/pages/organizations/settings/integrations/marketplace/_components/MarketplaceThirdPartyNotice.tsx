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

import { LightbulbIcon, SparkleIcon } from "@phosphor-icons/react";
import { ThirdPartyLogo } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { TonedCard } from "#/components/TonedCard/TonedCard";

interface MarketplaceThirdPartyNoticeProps {
  providers: readonly {
    provider: string;
    displayName: string;
  }[];
  onSuggest: () => void;
}

export function MarketplaceThirdPartyNotice({
  providers,
  onSuggest,
}: MarketplaceThirdPartyNoticeProps) {
  const { t } = useTranslation("organizations/settings/integrations");

  return (
    <TonedCard
      tone="sky"
      fill={false}
      className="w-full"
      icon={<LightbulbIcon weight="duotone" />}
      lead={(
        <Heading level={2} size={3} weight="medium" highContrast>
          {t("marketplacePage.thirdParties.title")}
        </Heading>
      )}
      control={(
        <Button
          size={2}
          variant="solid"
          color="neutral"
          highContrast
          iconStart={<SparkleIcon />}
          onClick={onSuggest}
        >
          {t("marketplacePage.thirdParties.filterAction")}
        </Button>
      )}
    >
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Text size={2} color="neutral" className="min-w-48 flex-1">
          {t("marketplacePage.thirdParties.description")}
        </Text>
        <ul
          aria-label={t("marketplacePage.thirdParties.title")}
          className="flex flex-wrap justify-end gap-2"
        >
          {providers.map(provider => (
            <li key={provider.provider} className="flex">
              <ThirdPartyLogo
                thirdParty={provider.provider}
                aria-label={provider.displayName}
                tint
                className="size-6 shrink-0 opacity-60"
              />
            </li>
          ))}
        </ul>
      </div>
    </TonedCard>
  );
}
