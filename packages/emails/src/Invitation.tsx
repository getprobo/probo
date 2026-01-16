import { Button, Section, Text } from '@react-email/components';
import * as React from 'react';
import EmailLayout, { bodyText, button, buttonContainer, footerText } from './components/EmailLayout';

export const Invitation = () => {
  return (
    <EmailLayout subject={'Invitation to join {{.OrganizationName}} on Probo'}>
      <Text style={bodyText}>
        You have been invited to join organization <strong>{'{{.OrganizationName}}'}</strong> on Probo. Click the button below to accept the invitation:
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href={'{{.InvitationUrl}}'}>
          Accept Invitation
        </Button>
      </Section>

      <Text style={footerText}>
        If you don't want to accept the invitation, you can ignore this email.
      </Text>
    </EmailLayout>
  );
};

export default Invitation;
