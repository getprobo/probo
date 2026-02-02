import { Button, Section, Text } from "@react-email/components";
import * as React from "react";
import EmailLayout, {
  bodyText,
  button,
  buttonContainer,
  footerText,
} from "./components/EmailLayout";

export const TrustCenterAccess = () => {
  return (
    <EmailLayout
      subject={`Compliance Page Access Invitation - ${"{{.OrganizationName}}"}`}
    >
      <Text style={bodyText}>
        You have been granted access to{" "}
        <strong>{"{{.OrganizationName}}"}</strong>'s compliance page! Click the
        button below to access it:
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href={"{{.BaseURL}}"}>
          Access Compliance Page
        </Button>
      </Section>
    </EmailLayout>
  );
};

export default TrustCenterAccess;
