import { Img } from '@react-email/components';
import * as React from 'react';

export function ProboLogo() {
  return (
    <Img
      style={{verticalAlign: "middle", height: "100%"}}
      src="{{.PoweredByLogoURL}}"
      alt="Probo"
      height="16"
    />
  );
}
