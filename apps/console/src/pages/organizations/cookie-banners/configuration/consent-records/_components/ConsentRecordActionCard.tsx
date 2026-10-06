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

import {
  CheckCircleIcon,
  EyeIcon,
  HandPalmIcon,
  SlidersHorizontalIcon,
  XCircleIcon,
} from "@phosphor-icons/react";
import { dateFormat } from "@probo/i18n";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ConsentRecordActionCard_cookieConsentRecord$key } from "#/__generated__/core/ConsentRecordActionCard_cookieConsentRecord.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import type { TonedCardTone } from "#/components/TonedCard/variants";

import { consentRecordPage } from "../../../variants";
import { getActionBadgeColor } from "../_lib/consentRecordHelpers";

import { ConsentRecordCopyValue } from "./ConsentRecordCopyValue";

const consentRecordActionCardFragment = graphql`
  fragment ConsentRecordActionCard_cookieConsentRecord on CookieConsentRecord {
    action
    visitorId
    createdAt
    cookieBannerVersion @required(action: THROW) {
      version
    }
  }
`;

function actionTone(action: string): TonedCardTone {
  const color = getActionBadgeColor(action);
  if (color === "neutral") {
    return "sand";
  }
  return color;
}

function actionIcon(action: string): ReactNode {
  switch (action) {
    case "ACCEPT_ALL":
      return <CheckCircleIcon size={24} weight="duotone" />;
    case "REJECT_ALL":
      return <XCircleIcon size={24} weight="duotone" />;
    case "CUSTOMIZE":
      return <SlidersHorizontalIcon size={24} weight="duotone" />;
    case "GPC":
      return <HandPalmIcon size={24} weight="duotone" />;
    default:
      return <EyeIcon size={24} weight="duotone" />;
  }
}

interface ConsentRecordActionCardProps {
  className?: string;
  cookieConsentRecordKey: ConsentRecordActionCard_cookieConsentRecord$key;
}

export function ConsentRecordActionCard({
  className,
  cookieConsentRecordKey,
}: ConsentRecordActionCardProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const record = useFragment(consentRecordActionCardFragment, cookieConsentRecordKey);
  const { fields, property, span } = consentRecordPage();

  return (
    <TonedCard
      className={className}
      tone={actionTone(record.action)}
      fill={false}
      icon={actionIcon(record.action)}
      lead={(
        <Heading level={2} size={4} weight="medium" highContrast>
          {t(`consentRecordPage.actions.${record.action.toLowerCase()}`)}
        </Heading>
      )}
    >
      <div className={fields()}>
        <div className={property()}>
          <Text size={1} color="faint">{t("consentRecordPage.properties.bannerVersion")}</Text>
          <Text size={2} highContrast>
            {t("configLayout.callout.version", { version: record.cookieBannerVersion.version })}
          </Text>
        </div>
        <div className={property()}>
          <Text size={1} color="faint">{t("consentRecordPage.properties.date")}</Text>
          <Text size={2} highContrast>
            <time dateTime={record.createdAt}>
              {dateFormat(i18n.language, record.createdAt)}
            </time>
          </Text>
        </div>
        <div className={property({ className: span() })}>
          <Text size={1} color="faint">{t("consentRecordPage.properties.visitorId")}</Text>
          <ConsentRecordCopyValue
            label={t("consentRecordPage.properties.visitorId")}
            value={record.visitorId}
          />
        </div>
      </div>
    </TonedCard>
  );
}
