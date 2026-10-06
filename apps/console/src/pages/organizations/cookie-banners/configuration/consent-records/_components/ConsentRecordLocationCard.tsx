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

import { MapPinIcon } from "@phosphor-icons/react";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { CardInset } from "@probo/ui/src/v2/Card/CardInset";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ConsentRecordLocationCard_cookieConsentRecord$key } from "#/__generated__/core/ConsentRecordLocationCard_cookieConsentRecord.graphql";

import { consentRecordPage } from "../../../variants";
import { formatLocation } from "../_lib/consentRecordHelpers";
import {
  isoNumericId,
  WORLD_MAP_HEIGHT,
  WORLD_MAP_WIDTH,
  worldCountryPaths,
} from "../_lib/worldMap";

const consentRecordLocationCardFragment = graphql`
  fragment ConsentRecordLocationCard_cookieConsentRecord on CookieConsentRecord {
    countryCode
    subdivisionCode
    regulation
  }
`;

interface ConsentRecordLocationCardProps {
  className?: string;
  cookieConsentRecordKey: ConsentRecordLocationCard_cookieConsentRecord$key;
}

export function ConsentRecordLocationCard({
  className,
  cookieConsentRecordKey,
}: ConsentRecordLocationCardProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const record = useFragment(consentRecordLocationCardFragment, cookieConsentRecordKey);
  const location = formatLocation(record.countryCode, record.subdivisionCode);
  const activeId = record.countryCode == null ? null : isoNumericId(record.countryCode);
  const showMap = record.countryCode != null;
  const { card, title, titleIcon, fields, property, map, mapSvg, country, countryActive } = consentRecordPage();

  return (
    <Card size={2} variant="soft" className={className}>
      {showMap
        ? (
            <CardInset side="top" className={map()}>
              <svg
                className={mapSvg()}
                viewBox={`0 0 ${WORLD_MAP_WIDTH} ${WORLD_MAP_HEIGHT}`}
                aria-hidden
              >
                {worldCountryPaths().map(path => (
                  <path
                    key={path.id}
                    d={path.d}
                    className={path.id === activeId ? countryActive() : country()}
                    fillRule="evenodd"
                    strokeWidth={0.5}
                  />
                ))}
              </svg>
            </CardInset>
          )
        : null}
      <div className={card()}>
        <div className={title()}>
          <MapPinIcon size={20} weight="duotone" className={titleIcon()} />
          <Heading level={2} size={4} weight="medium" highContrast>
            {t("consentRecordPage.cards.location")}
          </Heading>
        </div>
        <div className={fields()}>
          <div className={property()}>
            <Text size={1} color="faint">{t("consentRecordPage.cards.location")}</Text>
            <Text size={2} highContrast>
              {location === "" ? t("consentRecordPage.unknownLocation") : location}
            </Text>
          </div>
          <div className={property()}>
            <Text size={1} color="faint">{t("consentRecordPage.properties.regulation")}</Text>
            <Text size={2} highContrast>
              {record.regulation ?? "-"}
            </Text>
          </div>
        </div>
      </div>
    </Card>
  );
}
