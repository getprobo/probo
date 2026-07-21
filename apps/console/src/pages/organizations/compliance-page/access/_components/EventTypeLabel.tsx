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

import { useTranslation } from "react-i18next";

export function EventTypeLabel({ eventType }: { eventType: string }) {
  const { t } = useTranslation("organizations/compliance-page");

  switch (eventType) {
    case "DOCUMENT_VIEWED":
      return t("eventTypeLabel.documentViewed");
    case "CONSENT_GIVEN":
      return t("eventTypeLabel.consentGiven");
    case "FULL_NAME_TYPED":
      return t("eventTypeLabel.fullNameTyped");
    case "SIGNATURE_ACCEPTED":
      return t("eventTypeLabel.signatureAccepted");
    case "SIGNATURE_COMPLETED":
      return t("eventTypeLabel.signatureCompleted");
    case "SEAL_COMPUTED":
      return t("eventTypeLabel.sealComputed");
    case "TIMESTAMP_REQUESTED":
      return t("eventTypeLabel.timestampRequested");
    case "CERTIFICATE_GENERATED":
      return t("eventTypeLabel.certificateGenerated");
    case "PROCESSING_ERROR":
      return t("eventTypeLabel.processingError");
    default:
      return eventType;
  }
}
