import { Button, Section, Text } from "@react-email/components";
import * as React from "react";
import EmailLayout, {
  bodyText,
  button,
  buttonContainer,
  footerText,
} from "./components/EmailLayout";

export const MagicLink = () => {
  return (
    <EmailLayout subject="Probo Magic Link">
      <Text style={bodyText}>{"Please use this link to connect to {{.OrganizationName}}'s Compliance Page:"}</Text>

      <Section style={buttonContainer}>
        <Button style={button} href={"{{.MagicLinkURL}}"}>
          Connect
        </Button>
      </Section>

      <Text style={footerText}>
        This link will expire in {"{{.DurationInMinutes}}"} minutes.
      </Text>
    </EmailLayout>
  );
};

export default MagicLink;
