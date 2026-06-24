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
