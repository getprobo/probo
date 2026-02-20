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
