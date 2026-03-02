import { Text } from "@react-email/components";
import * as React from "react";
import EmailLayout, {
  bodyText,
  footerText,
} from "./components/EmailLayout";

export const ElectronicSignatureCertificate = () => {
  return (
    <EmailLayout
      subject={`Your signed ${"{{.DocumentType}}"} — Certificate of Completion`}
    >
      <Text style={bodyText}>
        Your <strong>{"{{.DocumentType}}"}</strong> has been signed
        electronically. A Certificate of Completion is attached to this email
        as a PDF document.
      </Text>

      <Text style={bodyText}>
        The certificate contains a complete record of the signing event,
        including the integrity seal, timestamp, and full audit trail.
      </Text>

      <Text style={footerText}>
        If you have any questions, please contact support.
      </Text>
    </EmailLayout>
  );
};

export default ElectronicSignatureCertificate;
