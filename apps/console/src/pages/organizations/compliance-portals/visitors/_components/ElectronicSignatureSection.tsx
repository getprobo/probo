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

import { DownloadSimpleIcon, InfoIcon, SignatureIcon } from "@phosphor-icons/react";
import { safeOpenUrl } from "@probo/helpers";
import { dateTimeFormat } from "@probo/i18n";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Popover } from "@probo/ui/src/v2/Popover/Popover";
import { PopoverPopup } from "@probo/ui/src/v2/Popover/PopoverPopup";
import { PopoverTrigger } from "@probo/ui/src/v2/Popover/PopoverTrigger";
import { Timeline } from "@probo/ui/src/v2/Timeline/Timeline";
import { TimelineContent } from "@probo/ui/src/v2/Timeline/TimelineContent";
import { TimelineItem } from "@probo/ui/src/v2/Timeline/TimelineItem";
import { TimelineMarker } from "@probo/ui/src/v2/Timeline/TimelineMarker";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ElectronicSignatureSectionFragment$key } from "#/__generated__/core/ElectronicSignatureSectionFragment.graphql";
import { CompliancePortalTonedCard } from "#/pages/organizations/compliance-portals/_components/CompliancePortalTonedCard";

import { type NdaSignatureStatus, ndaSignatureTone } from "../_lib/ndaSignature";
import { electronicSignatureSection } from "../variants";

import { EventTypeIcon } from "./EventTypeIcon";
import { EventTypeLabel } from "./EventTypeLabel";

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
  const { root, copy, description, trigger, popup, event, timestamp }
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
    <section className={root()} aria-label={t("electronicSignature.title")}>
      <CompliancePortalTonedCard
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
              <div className={copy()}>
                <Text size={2} color="neutral" className={description()}>
                  {descriptionCopy}
                </Text>
                <Popover>
                  <PopoverTrigger
                    render={(
                      <IconButton
                        size={1}
                        variant="ghost"
                        color="neutral"
                        aria-label={t("electronicSignature.activity")}
                        className={trigger()}
                      />
                    )}
                  >
                    <InfoIcon />
                  </PopoverTrigger>
                  <PopoverPopup className={popup()} align="end">
                    <Timeline>
                      {signature.events.map(item => (
                        <TimelineItem key={item.id}>
                          <TimelineMarker
                            color={item.eventType === "PROCESSING_ERROR" ? "red" : "neutral"}
                          >
                            <EventTypeIcon eventType={item.eventType} />
                          </TimelineMarker>
                          <TimelineContent>
                            <Text size={1} highContrast className={event()}>
                              <EventTypeLabel eventType={item.eventType} />
                            </Text>
                            <Text size={1} color="faint" className={timestamp()}>
                              {dateTimeFormat(i18n.language, item.occurredAt, signedAtFormat)}
                            </Text>
                          </TimelineContent>
                        </TimelineItem>
                      ))}
                    </Timeline>
                  </PopoverPopup>
                </Popover>
              </div>
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
