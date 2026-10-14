// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
import EmailLayout, { bodyText, button, buttonContainer } from './components/EmailLayout';

export const DeviceInstallation = () => {
  return (
    <EmailLayout subject={'Install the Probo device agent for {{.OrganizationName}}'}>
      <Text style={bodyText}>
        <strong>{'{{.OrganizationName}}'}</strong>
        {' uses Probo to keep a record of the security posture of each work device, the computer an employee uses for '}
        <strong>{'{{.OrganizationName}}'}</strong>
        {'. The device agent checks settings such as disk encryption and screen lock, then reports the result to '}
        <strong>{'{{.OrganizationName}}'}</strong>
        {'. Install it on the computer you use for '}
        <strong>{'{{.OrganizationName}}'}</strong>
        {' so that computer is included in the record.'}
      </Text>

      <Text style={bodyText}>
        The agent collects configuration data only, and nothing else. It never collects personal data.
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href={'{{.InstallUrl}}'}>
          Install the device agent
        </Button>
      </Section>
    </EmailLayout>
  );
};

export default DeviceInstallation;
