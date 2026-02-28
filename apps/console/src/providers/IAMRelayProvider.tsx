import type { ReactNode } from "react";
import { RelayEnvironmentProvider } from "react-relay";

import { iamEnvironment } from "#/environments";

export function ConsoleRelayProvider(props: { children: ReactNode }) {
  return <RelayEnvironmentProvider environment={iamEnvironment} {...props} />;
}

// Legacy export for backward compatibility
export function IAMRelayProvider(props: { children: ReactNode }) {
  return <ConsoleRelayProvider {...props} />;
}
