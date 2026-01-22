import type { ReactNode } from "react";
import { RelayEnvironmentProvider } from "react-relay";

import { iamEnvironment } from "#/environments";

export function IAMRelayProvider(props: { children: ReactNode }) {
  return <RelayEnvironmentProvider environment={iamEnvironment} {...props} />;
}
