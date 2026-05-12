// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";

export function EventTypeLabel({ eventType }: { eventType: string }) {
  const { __ } = useTranslate();

  switch (eventType) {
    case "DOCUMENT_VIEWED":
      return __("Document viewed");
    case "CONSENT_GIVEN":
      return __("Consent given");
    case "FULL_NAME_TYPED":
      return __("Full name typed");
    case "SIGNATURE_ACCEPTED":
      return __("Signature accepted");
    case "SIGNATURE_COMPLETED":
      return __("Signature completed");
    case "SEAL_COMPUTED":
      return __("Seal computed");
    case "TIMESTAMP_REQUESTED":
      return __("Timestamp requested");
    case "CERTIFICATE_GENERATED":
      return __("Certificate generated");
    case "PROCESSING_ERROR":
      return __("Processing error");
    default:
      return eventType;
  }
}
