import { Button, Section, Text } from '@react-email/components';
import * as React from 'react';
import EmailLayout, { bodyText, button, buttonContainer, footerText } from './components/EmailLayout';

export const ConfirmEmail = () => {
  return (
    <EmailLayout subject="Confirm your email address">
      <Text style={bodyText}>
        Thanks for joining Probo! Please confirm your email address by clicking the button below:
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href={'{{.ConfirmationUrl}}'}>
          Confirm Email Address
        </Button>
      </Section>

      <Text style={footerText}>
        If you did not sign up for Probo, please ignore this email.
      </Text>
    </EmailLayout>
  );
};

export default ConfirmEmail;
