import { Button, Hr, Link, Section, Text } from "@react-email/components";
import * as React from "react";
import EmailLayout, {
  bodyText,
  button,
  buttonContainer,
  footerText,
} from "./components/EmailLayout";

export const MailingListSubscription = () => {
  return (
    <EmailLayout
      subject={`${"{{.OrganizationName}}"} – Confirm Your Compliance Updates Subscription`}
    >
      <Text style={bodyText}>
        You requested to subscribe to compliance updates from{" "}
        <strong>{"{{.OrganizationName}}"}</strong>.
      </Text>

      <Text style={bodyText}>
        Please click the button below to confirm your subscription.
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href={"{{.ConfirmURL}}"}>
          Confirm Subscription
        </Button>
      </Section>

      <Hr style={{ borderColor: "#ecefec", margin: "8px 0 20px" }} />

      <Text style={{ ...footerText, textAlign: "center" }}>
        If you did not request this subscription, or no longer want to receive
        these emails,{" "}
        <Link
          href={"{{.UnsubscribeURL}}"}
          style={{ color: "#374151", textDecoration: "underline", fontWeight: "500" }}
        >
          unsubscribe here.
        </Link>
      </Text>
    </EmailLayout>
  );
};

export default MailingListSubscription;
