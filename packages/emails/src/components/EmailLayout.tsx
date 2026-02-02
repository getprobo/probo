import {
  Body,
  Container,
  Head,
  Html,
  Link,
  Section,
  Text,
} from '@react-email/components';
import * as React from 'react';
import { Logo } from './Logo';
import { ProboLogo } from './ProboLogo';

interface EmailLayoutProps {
  subject: string;
  children: React.ReactNode;
}

export const EmailLayout = ({
  subject,
  children,
}: EmailLayoutProps) => {
  return (
    <Html lang="en">
      <Head>
        <meta name="x-apple-disable-message-reformatting" />
        <meta httpEquiv="X-UA-Compatible" content="IE=edge" />
        <title>{subject}</title>
      </Head>
      <Body style={main}>
        <Container style={container}>
          <Section style={content}>
            <Section style={logoSection}>
              <Link href="{{.SenderCompanyWebsiteURL}}">
                <Logo />
              </Link>
            </Section>

            <Text style={text}>Hi {'{{.RecipientFullName}}'},</Text>

            {children}
          </Section>

          <Section style={footerSection}>
            <Text style={footerAddress}>
              {"{{.SenderCompanyHeadquarterAddress}}"}
            </Text>
            <Text style={footerAddress}>
              <span style={{verticalAlign: "middle"}}>Powered By </span>
              <Link style={{display: "inline-block", height: "16px", verticalAlign: "middle"}} href="https://www.getprobo.com">
                <ProboLogo />
              </Link>
            </Text>
          </Section>
        </Container>
      </Body>
    </Html>
  );
};

export default EmailLayout;

const main: React.CSSProperties = {
  margin: '0',
  padding: '0',
  fontFamily:
    "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
  backgroundColor: '#fffffb',
  WebkitTextSizeAdjust: '100%',
};

const container: React.CSSProperties = {
  maxWidth: '600px',
  width: '100%',
  backgroundColor: '#ffffff',
  borderRadius: '8px',
  boxShadow: '0 2px 4px rgba(0, 0, 0, 0.1)',
  margin: '40px auto',
};

export const headerSection: React.CSSProperties = {
  padding: '40px 40px 30px 40px',
  textAlign: 'center',
  backgroundColor: '#101e1c',
  borderRadius: '8px 8px 0 0',
};

export const h1: React.CSSProperties = {
  margin: '0',
  color: '#e6ff03',
  fontSize: '24px',
  fontWeight: '600',
  lineHeight: '1.3',
};

const content: React.CSSProperties = {
  padding: '40px',
};

const text: React.CSSProperties = {
  margin: '0 0 20px 0',
  color: '#141e12',
  fontSize: '16px',
  lineHeight: '24px',
};

export const buttonContainer: React.CSSProperties = {
  padding: '10px 0 30px 0',
};

export const button: React.CSSProperties = {
  display: 'inline-block',
  padding: '14px 28px',
  backgroundColor: '#101e1c',
  color: '#ffffff',
  textDecoration: 'none',
  borderRadius: '8px',
  fontWeight: '600',
  fontSize: '16px',
  lineHeight: '20px',
};

export const footerText: React.CSSProperties = {
  margin: '0',
  color: '#6b716a',
  fontSize: '14px',
  lineHeight: '20px',
};

const footerSection: React.CSSProperties = {
  padding: '30px 40px',
  borderTop: '1px solid #ecefec',
};

const footerAddress: React.CSSProperties = {
  margin: '10px 0 0 0',
  color: '#6b716a',
  fontSize: '12px',
  lineHeight: '18px',
};

const logoSection: React.CSSProperties = {
  marginBottom: '30px',
};

export const bodyText: React.CSSProperties = {
  margin: '0 0 30px 0',
  color: '#141e12',
  fontSize: '16px',
  lineHeight: '24px',
};
