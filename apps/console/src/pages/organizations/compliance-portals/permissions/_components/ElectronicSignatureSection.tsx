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

import { CaretDownIcon, DownloadSimpleIcon, SignatureIcon } from "@phosphor-icons/react";
import { safeOpenUrl } from "@probo/helpers";
import { dateTimeFormat } from "@probo/i18n";
import { Collapsible } from "@probo/ui/src/v2/Collapsible/Collapsible";
import { CollapsiblePanel } from "@probo/ui/src/v2/Collapsible/CollapsiblePanel";
import { CollapsibleTrigger } from "@probo/ui/src/v2/Collapsible/CollapsibleTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Timeline } from "@probo/ui/src/v2/Timeline/Timeline";
import { TimelineContent } from "@probo/ui/src/v2/Timeline/TimelineContent";
import { TimelineItem } from "@probo/ui/src/v2/Timeline/TimelineItem";
import { TimelineMarker } from "@probo/ui/src/v2/Timeline/TimelineMarker";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ElectronicSignatureSectionFragment$key } from "#/__generated__/core/ElectronicSignatureSectionFragment.graphql";
import { CompliancePortalTonedCard } from "#/pages/organizations/compliance-portals/_components/CompliancePortalTonedCard";

import { electronicSignatureSection } from "../variants";

import { EventTypeIcon } from "./EventTypeIcon";
import { EventTypeLabel } from "./EventTypeLabel";
import type { NdaSignatureStatus } from "./NdaSignatureBadge";
import { ndaSignatureTone } from "./NdaSignatureBadge";

const signedAtFormat = {
  hour: "2-digit",
  hour12: false,
  minute: "2-digit",
  second: "2-digit",
  day: "numeric",
  month: "short",
  year: "numeric",
} as const;

const fragment = graphql`
  fragment ElectronicSignatureSectionFragment on ElectronicSignature {
    status
    signedAt
    certificate {
      downloadUrl
    }
    events {
      id
      eventType
      occurredAt
    }
  }
`;

function signatureLeadKey(status: NdaSignatureStatus): string {
  switch (status) {
    case "COMPLETED":
      return "electronicSignature.status.signed";
    case "ACCEPTED":
    case "PROCESSING":
      return "electronicSignature.status.processing";
    case "PENDING":
      return "electronicSignature.status.pending";
    case "FAILED":
      return "electronicSignature.status.failed";
  }
}

function signatureDescription(
  status: NdaSignatureStatus,
  signedAt: string | null | undefined,
  language: string,
  t: (key: string, options?: { date: string }) => string,
): string {
  if (signedAt != null) {
    return t("electronicSignature.descriptions.signedAt", {
      date: dateTimeFormat(language, signedAt, signedAtFormat),
    });
  }
  switch (status) {
    case "PENDING":
      return t("electronicSignature.descriptions.pending");
    case "ACCEPTED":
    case "PROCESSING":
      return t("electronicSignature.descriptions.processing");
    case "FAILED":
      return t("electronicSignature.descriptions.failed");
    case "COMPLETED":
      return t("electronicSignature.descriptions.completed");
  }
}

export function ElectronicSignatureSection({
  fragmentRef,
}: {
  fragmentRef: ElectronicSignatureSectionFragment$key;
}) {
  const { i18n, t } = useTranslation("organizations/compliance-portals");
  const signature = useFragment(fragment, fragmentRef);
  const { root, card, activity, copy, description, trigger }
    = electronicSignatureSection();
  const tone = ndaSignatureTone(signature.status);
  const certificateUrl = signature.certificate?.downloadUrl;
  const descriptionCopy = signatureDescription(
    signature.status,
    signature.signedAt,
    i18n.language,
    t,
  );

  return (
    <section className={root()}>
      <Heading level={3} size={3} weight="medium" highContrast>
        {t("electronicSignature.title")}
      </Heading>
      <CompliancePortalTonedCard
        className={card()}
        tone={tone}
        icon={<SignatureIcon size={24} weight="duotone" />}
        lead={(
          <Text size={2} weight="medium" color={tone} className="truncate">
            {t(signatureLeadKey(signature.status))}
          </Text>
        )}
        control={certificateUrl != null
          ? (
              <IconButton
                size={1}
                variant="ghost"
                color="neutral"
                aria-label={t("electronicSignature.actions.download")}
                onClick={() => {
                  safeOpenUrl(certificateUrl);
                }}
              >
                <DownloadSimpleIcon />
              </IconButton>
            )
          : undefined}
      >
        <Text size={3} weight="medium" highContrast>
          {t("electronicSignature.certificateTitle")}
        </Text>
        {signature.events.length > 0
          ? (
              <Collapsible className={activity()}>
                <div className={copy()}>
                  <Text size={2} color="neutral" className={description()}>
                    {descriptionCopy}
                  </Text>
                  <CollapsibleTrigger
                    className={trigger()}
                    render={(
                      <IconButton
                        size={1}
                        variant="ghost"
                        color="neutral"
                        aria-label={t("electronicSignature.activity")}
                      />
                    )}
                  >
                    <CaretDownIcon />
                  </CollapsibleTrigger>
                </div>
                <CollapsiblePanel>
                  <Timeline>
                    {signature.events.map(item => (
                      <TimelineItem key={item.id}>
                        <TimelineMarker
                          color={item.eventType === "PROCESSING_ERROR" ? "red" : "neutral"}
                        >
                          <EventTypeIcon eventType={item.eventType} />
                        </TimelineMarker>
                        <TimelineContent>
                          <Text size={1} highContrast>
                            <EventTypeLabel eventType={item.eventType} />
                          </Text>
                          <Text size={1} color="faint">
                            {dateTimeFormat(i18n.language, item.occurredAt, signedAtFormat)}
                          </Text>
                        </TimelineContent>
                      </TimelineItem>
                    ))}
                  </Timeline>
                </CollapsiblePanel>
              </Collapsible>
            )
          : (
              <Text size={2} color="neutral">
                {descriptionCopy}
              </Text>
            )}
      </CompliancePortalTonedCard>
    </section>
  );
}
