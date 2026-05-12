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
