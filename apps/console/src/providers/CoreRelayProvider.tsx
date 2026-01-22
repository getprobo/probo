import type { ReactNode } from "react";
import { RelayEnvironmentProvider } from "react-relay";

import { coreEnvironment } from "#/environments";

export function CoreRelayProvider(props: { children: ReactNode }) {
  return <RelayEnvironmentProvider environment={coreEnvironment} {...props} />;
}
