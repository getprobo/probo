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

import { safeOpenUrl } from "@probo/helpers";
import { dateTimeFormat } from "@probo/i18n";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ElectronicSignatureSectionFragment$key } from "#/__generated__/core/ElectronicSignatureSectionFragment.graphql";

import { electronicSignatureSection } from "../variants";

import { EventTypeLabel } from "./EventTypeLabel";
import { NdaSignatureBadge } from "./NdaSignatureBadge";

const fragment = graphql`
  fragment ElectronicSignatureSectionFragment on ElectronicSignature {
    status
    signedAt
    certificate {
      downloadUrl
      fileName
    }
    events {
      id
      eventType
      actorEmail
      occurredAt
    }
  }
`;

export function ElectronicSignatureSection({
  fragmentRef,
}: {
  fragmentRef: ElectronicSignatureSectionFragment$key;
}) {
  const { i18n, t } = useTranslation("organizations/compliance-portals");
  const signature = useFragment(fragment, fragmentRef);
  const { root, rows, row, activity, event, eventCopy } = electronicSignatureSection();
  const certificateUrl = signature.certificate?.downloadUrl;

  return (
    <section className={root()}>
      <Heading level={3} size={3} weight="medium" highContrast>
        {t("electronicSignature.title")}
      </Heading>
      <Card variant="soft" size={2}>
        <div className={rows()}>
          <div className={row()}>
            <Text size={2} color="neutral">
              {t("electronicSignature.fields.status")}
            </Text>
            <NdaSignatureBadge status={signature.status} />
          </div>
          {signature.signedAt != null && (
            <div className={row()}>
              <Text size={2} color="neutral">
                {t("electronicSignature.fields.signedAt")}
              </Text>
              <Text size={2} highContrast>
                {dateTimeFormat(i18n.language, signature.signedAt)}
              </Text>
            </div>
          )}
          {certificateUrl != null && (
            <div className={row()}>
              <Text size={2} color="neutral">
                {t("electronicSignature.fields.certificate")}
              </Text>
              <Button
                type="button"
                variant="ghost"
                color="neutral"
                size={1}
                onClick={() => {
                  safeOpenUrl(certificateUrl);
                }}
              >
                {signature.certificate?.fileName ?? t("electronicSignature.actions.download")}
              </Button>
            </div>
          )}
          {signature.events.length > 0 && (
            <div className={activity()}>
              <Text size={1} weight="medium" color="faint">
                {t("electronicSignature.activity")}
              </Text>
              {signature.events.map(item => (
                <div key={item.id} className={event()}>
                  <div className={eventCopy()}>
                    <Text size={1} highContrast>
                      <EventTypeLabel eventType={item.eventType} />
                    </Text>
                    {item.actorEmail != null && (
                      <Text size={1} color="faint">
                        {item.actorEmail}
                      </Text>
                    )}
                  </div>
                  <Text size={1} color="faint">
                    {dateTimeFormat(i18n.language, item.occurredAt)}
                  </Text>
                </div>
              ))}
            </div>
          )}
        </div>
      </Card>
    </section>
  );
}
