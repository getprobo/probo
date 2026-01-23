import relay from "eslint-plugin-relay";
import type { FlatConfig } from "typescript-eslint";

export const relayConfigs: FlatConfig.ConfigArray = [
  {
    plugins: {
      relay: relay,
    },
    ...relay.configs["ts-recommended"],
  },
];
