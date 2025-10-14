import { Button, Section, Text } from '@react-email/components';
import * as React from 'react';
import EmailLayout, { bodyText, button, buttonContainer, footerText } from './components/EmailLayout';

export const PasswordReset = () => {
  return (
    <EmailLayout subject="Reset your password">
      <Text style={bodyText}>
        You have requested a password reset for your Probo account. Click the button below to reset your password:
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href={'{{.ResetUrl}}'}>
          Reset Password
        </Button>
      </Section>

      <Text style={footerText}>
        If you did not request this password reset, please ignore this email.
      </Text>
    </EmailLayout>
  );
};

export default PasswordReset;
