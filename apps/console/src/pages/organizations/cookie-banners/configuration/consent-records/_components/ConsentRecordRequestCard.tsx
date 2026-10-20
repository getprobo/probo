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

import { BrowserIcon } from "@phosphor-icons/react";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ConsentRecordRequestCard_cookieConsentRecord$key } from "#/__generated__/core/ConsentRecordRequestCard_cookieConsentRecord.graphql";

import { consentRecordPage } from "../../../variants";
import { formatAnonymizedIp } from "../_lib/consentRecordHelpers";

import { ConsentRecordCopyValue } from "./ConsentRecordCopyValue";

const consentRecordRequestCardFragment = graphql`
  fragment ConsentRecordRequestCard_cookieConsentRecord on CookieConsentRecord {
    origin
    ipAddress
    userAgent
    sdkVersion
  }
`;

interface ConsentRecordRequestCardProps {
  className?: string;
  cookieConsentRecordKey: ConsentRecordRequestCard_cookieConsentRecord$key;
}

function requestValue(value: string | null | undefined): string | null {
  if (value == null || value === "") {
    return null;
  }
  return value;
}

export function ConsentRecordRequestCard({
  className,
  cookieConsentRecordKey,
}: ConsentRecordRequestCardProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const record = useFragment(consentRecordRequestCardFragment, cookieConsentRecordKey);
  const { card, title, titleIcon, fields, property, span } = consentRecordPage();
  const ip = record.ipAddress == null || record.ipAddress === ""
    ? null
    : formatAnonymizedIp(record.ipAddress);

  return (
    <Card size={2} variant="soft" className={className}>
      <div className={card()}>
        <div className={title()}>
          <BrowserIcon size={20} weight="duotone" className={titleIcon()} />
          <Heading level={2} size={4} weight="medium" highContrast>
            {t("consentRecordPage.cards.request")}
          </Heading>
        </div>
        <div className={fields()}>
          <div className={property({ className: span() })}>
            <Text size={1} color="faint">{t("consentRecordPage.properties.origin")}</Text>
            <ConsentRecordCopyValue
              label={t("consentRecordPage.properties.origin")}
              value={requestValue(record.origin)}
            />
          </div>
          <div className={property()}>
            <Text size={1} color="faint">{t("consentRecordPage.properties.ipAddress")}</Text>
            <ConsentRecordCopyValue
              label={t("consentRecordPage.properties.ipAddress")}
              value={ip}
            />
          </div>
          <div className={property()}>
            <Text size={1} color="faint">{t("consentRecordPage.properties.sdkVersion")}</Text>
            <ConsentRecordCopyValue
              label={t("consentRecordPage.properties.sdkVersion")}
              value={requestValue(record.sdkVersion)}
            />
          </div>
          <div className={property({ className: span() })}>
            <Text size={1} color="faint">{t("consentRecordPage.properties.userAgent")}</Text>
            <ConsentRecordCopyValue
              label={t("consentRecordPage.properties.userAgent")}
              value={requestValue(record.userAgent)}
            />
          </div>
        </div>
      </div>
    </Card>
  );
}
