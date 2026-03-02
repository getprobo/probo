import { useTranslate } from "@probo/i18n";
import { Badge } from "@probo/ui";

type ElectronicSignatureStatus = "PENDING" | "ACCEPTED" | "PROCESSING" | "COMPLETED" | "FAILED";

export function NdaSignatureBadge({ status }: { status: ElectronicSignatureStatus }) {
  const { __ } = useTranslate();

  switch (status) {
    case "COMPLETED":
      return <Badge variant="success">{__("Signed")}</Badge>;
    case "ACCEPTED":
    case "PROCESSING":
      return <Badge variant="info">{__("Processing")}</Badge>;
    case "PENDING":
      return <Badge variant="warning">{__("Pending")}</Badge>;
    case "FAILED":
      return <Badge variant="danger">{__("Failed")}</Badge>;
  }
}
