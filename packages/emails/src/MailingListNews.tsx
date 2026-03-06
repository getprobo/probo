import { Button, Hr, Link, Section, Text } from "@react-email/components";
import * as React from "react";
import EmailLayout, {
  bodyText,
  button,
  buttonContainer,
  footerText,
} from "./components/EmailLayout";

export const MailingListNews = () => {
  return (
    <EmailLayout
      subject={`${"{{.OrganizationName}}"} – ${"{{.NewsTitle}}"}`}
    >
      <Text style={{ ...bodyText, fontWeight: "600", fontSize: "18px" }}>
        {"{{.NewsTitle}}"}
      </Text>

      <Text style={bodyText}>{"{{.NewsBody}}"}</Text>

      <Section style={buttonContainer}>
        <Button style={button} href={"{{.CompliancePageURL}}"}>
          View Compliance Page
        </Button>
      </Section>

      <Hr style={{ borderColor: "#ecefec", margin: "8px 0 20px" }} />

      <Text style={{ ...footerText, textAlign: "center" }}>
        You are receiving this email because you subscribed to compliance
        updates from <strong>{"{{.OrganizationName}}"}</strong>. To
        unsubscribe,{" "}
        <Link
          href={"{{.UnsubscribeURL}}"}
          style={{ color: "#374151", textDecoration: "underline", fontWeight: "500" }}
        >
          click here.
        </Link>
      </Text>
    </EmailLayout>
  );
};

export default MailingListNews;
