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

import { SquaresFourIcon } from "@phosphor-icons/react";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ConsentRecordCategoriesCard_cookieConsentRecord$key } from "#/__generated__/core/ConsentRecordCategoriesCard_cookieConsentRecord.graphql";

import { consentRecordPage } from "../../../variants";

const consentRecordCategoriesCardFragment = graphql`
  fragment ConsentRecordCategoriesCard_cookieConsentRecord on CookieConsentRecord {
    action
    consentData
    cookieBannerVersion @required(action: THROW) {
      categories {
        name
        slug
      }
    }
  }
`;

function parseConsentMap(consentData: string): Record<string, boolean> {
  try {
    const parsed: unknown = JSON.parse(consentData);
    if (parsed == null || typeof parsed !== "object" || Array.isArray(parsed)) {
      return {};
    }
    return parsed as Record<string, boolean>;
  } catch {
    return {};
  }
}

interface ConsentRecordCategoriesCardProps {
  className?: string;
  cookieConsentRecordKey: ConsentRecordCategoriesCard_cookieConsentRecord$key;
}

export function ConsentRecordCategoriesCard({
  className,
  cookieConsentRecordKey,
}: ConsentRecordCategoriesCardProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const record = useFragment(consentRecordCategoriesCardFragment, cookieConsentRecordKey);
  const consentMap = useMemo(() => parseConsentMap(record.consentData), [record.consentData]);
  const { card, title, titleIcon, category } = consentRecordPage();

  if (record.action !== "CUSTOMIZE") {
    return null;
  }

  return (
    <Card size={2} variant="soft" className={className}>
      <div className={card()}>
        <div className={title()}>
          <SquaresFourIcon size={20} weight="duotone" className={titleIcon()} />
          <Heading level={2} size={4} weight="medium" highContrast>
            {t("consentRecordPage.cards.categories")}
          </Heading>
        </div>
        {record.cookieBannerVersion.categories.map((item) => {
          const consented = consentMap[item.slug] === true;
          return (
            <div key={item.slug} className={category()}>
              <Text size={2} highContrast>{item.name}</Text>
              <Badge variant="soft" color={consented ? "green" : "red"}>
                {consented
                  ? t("consentRecordPage.consent.accepted")
                  : t("consentRecordPage.consent.rejected")}
              </Badge>
            </div>
          );
        })}
      </div>
    </Card>
  );
}
