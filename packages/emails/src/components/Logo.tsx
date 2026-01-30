import { Img } from '@react-email/components';
import * as React from 'react';

export function Logo() {
  return (
    <Img
      style={{ maxWidth: "220px" }}
      src="{{.SenderCompanyLogoURL}}"
      alt="{{.SenderCompanyName}}"
      height="60"
    />
  );
}
