import { Hr, Text } from "@react-email/components";
import * as React from "react";
import EmailLayout, { bodyText, footerText } from "./components/EmailLayout";

export const MailingListUnsubscription = () => {
  return (
    <EmailLayout
      subject={`${"{{.OrganizationName}}"} – You've been unsubscribed`}
    >
      <Text style={bodyText}>
        You've been successfully unsubscribed from{" "}
        <strong>{"{{.OrganizationName}}"}</strong>'s compliance updates. You
        will no longer receive notifications when new information is published
        on their compliance page.
      </Text>

      <Hr style={{ borderColor: "#ecefec", margin: "8px 0 20px" }} />

      <Text style={{ ...footerText, textAlign: "center" }}>
        This email was sent to confirm your unsubscription request.
      </Text>
    </EmailLayout>
  );
};

export default MailingListUnsubscription;
