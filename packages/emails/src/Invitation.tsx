// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { Button, Section, Text } from '@react-email/components';
import * as React from 'react';
import EmailLayout, { bodyText, button, buttonContainer, footerText } from './components/EmailLayout';

export const Invitation = () => {
  return (
    <EmailLayout subject={'Invitation to join {{.OrganizationName}} on Probo'}>
      <Text style={bodyText}>
        You have been invited to join organization <strong>{'{{.OrganizationName}}'}</strong> on Probo. Click the button below to activate your account:
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href={'{{.InvitationUrl}}'}>
          Activate Account
        </Button>
      </Section>

      <Text style={footerText}>
        If you don't want to do so, you can ignore this email.
      </Text>
    </EmailLayout>
  );
};

export default Invitation;
