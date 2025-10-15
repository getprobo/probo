import { Img } from '@react-email/components';
import * as React from 'react';

export function ProboLogo() {
  return (
    <Img
      src={'{{.LogoURL}}'}
      alt="Probo"
      width="220"
      height="60"
    />
  );
}
