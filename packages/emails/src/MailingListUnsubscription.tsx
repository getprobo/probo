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
