// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { Button, Section, Text } from 'react-email';
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
