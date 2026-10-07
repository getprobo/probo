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

import { SealCheckIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ConsentRecordTcfCard_cookieConsentRecord$key } from "#/__generated__/core/ConsentRecordTcfCard_cookieConsentRecord.graphql";

import { consentRecordPage } from "../../../variants";
import { decodeTcString } from "../_lib/decodeTcString";

import { ConsentRecordCopyValue } from "./ConsentRecordCopyValue";

const consentRecordTcfCardFragment = graphql`
  fragment ConsentRecordTcfCard_cookieConsentRecord on CookieConsentRecord {
    tc
  }
`;

interface ConsentRecordTcfCardProps {
  className?: string;
  cookieConsentRecordKey: ConsentRecordTcfCard_cookieConsentRecord$key;
}

export function ConsentRecordTcfCard({
  className,
  cookieConsentRecordKey,
}: ConsentRecordTcfCardProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const record = useFragment(consentRecordTcfCardFragment, cookieConsentRecordKey);
  const decoded = useMemo(
    () => record.tc == null || record.tc === "" ? null : decodeTcString(record.tc),
    [record.tc],
  );
  const { card, title, titleIcon, fields, property, span } = consentRecordPage();

  if (record.tc == null || record.tc === "") {
    return null;
  }

  return (
    <Card size={2} variant="soft" className={className}>
      <div className={card()}>
        <div className={title()}>
          <SealCheckIcon size={20} weight="duotone" className={titleIcon()} />
          <Heading level={2} size={4} weight="medium" highContrast>
            {t("consentRecordPage.cards.tcf")}
          </Heading>
        </div>
        <ConsentRecordCopyValue
          label={t("consentRecordPage.cards.tcf")}
          value={record.tc}
        />
        {decoded == null
          ? (
              <Text size={2} color="red">
                {t("consentRecordPage.tcf.decodeError")}
              </Text>
            )
          : (
              <div className={fields()}>
                <div className={property()}>
                  <Text size={1} color="faint">{t("consentRecordPage.tcf.created")}</Text>
                  <Text size={2} highContrast>
                    {dateTimeFormat(i18n.language, decoded.created)}
                  </Text>
                </div>
                <div className={property()}>
                  <Text size={1} color="faint">{t("consentRecordPage.tcf.lastUpdated")}</Text>
                  <Text size={2} highContrast>
                    {dateTimeFormat(i18n.language, decoded.lastUpdated)}
                  </Text>
                </div>
                <div className={property()}>
                  <Text size={1} color="faint">{t("consentRecordPage.tcf.cmp")}</Text>
                  <Text size={2} highContrast>
                    {t("consentRecordPage.tcf.cmpValue", {
                      id: decoded.cmpId,
                      version: decoded.cmpVersion,
                    })}
                  </Text>
                </div>
                <div className={property()}>
                  <Text size={1} color="faint">{t("consentRecordPage.tcf.language")}</Text>
                  <Text size={2} highContrast>{decoded.language || "-"}</Text>
                </div>
                <div className={property()}>
                  <Text size={1} color="faint">{t("consentRecordPage.tcf.tcfPolicyVersion")}</Text>
                  <Text size={2} highContrast>{decoded.tcfPolicyVersion}</Text>
                </div>
                <div className={property()}>
                  <Text size={1} color="faint">{t("consentRecordPage.tcf.gvlVersion")}</Text>
                  <Text size={2} highContrast>{decoded.gvlVersion}</Text>
                </div>
                <div className={property()}>
                  <Text size={1} color="faint">{t("consentRecordPage.tcf.vendorConsentCount")}</Text>
                  <Text size={2} highContrast>{decoded.vendorConsentCount}</Text>
                </div>
                <div className={property({ className: span() })}>
                  <Text size={1} color="faint">{t("consentRecordPage.tcf.purposes")}</Text>
                  <Text size={2} highContrast>
                    {decoded.purposeIds.length === 0 ? "-" : decoded.purposeIds.join(", ")}
                  </Text>
                </div>
              </div>
            )}
      </div>
    </Card>
  );
}
