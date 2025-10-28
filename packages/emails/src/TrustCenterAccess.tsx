import { Button, Section, Text } from '@react-email/components';
import * as React from 'react';
import EmailLayout, { bodyText, button, buttonContainer, footerText } from './components/EmailLayout';

export const TrustCenterAccess = () => {
  return (
    <EmailLayout subject={`Trust Center Access Invitation - ${'{{.OrganizationName}}'}`} organizationName={'{{.OrganizationName}}'}>
      <Text style={bodyText}>
        You have been granted access to <strong>{'{{.OrganizationName}}'}</strong>'s Trust Center! Click the button below to access it:
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href={'{{.AccessUrl}}'}>
          Access Trust Center
        </Button>
      </Section>

      <Text style={footerText}>
        This link will expire in {'{{.DurationInDays}}'} days.
      </Text>
    </EmailLayout>
  );
};

export default TrustCenterAccess;
