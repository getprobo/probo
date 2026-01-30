import { Img } from '@react-email/components';
import * as React from 'react';

export function Logo() {
  return (
    <Img
      className="max-width-[220px]"
      src="{{.SenderCompanyLogoURL}}"
      alt="{{.SenderCompanyName}}"
      height="60"
    />
  );
}
