import { makeFetchQuery } from "@probo/relay";
import { Environment, Network, RecordSource, Store } from "relay-runtime";

export const coreEnvironment = new Environment({
  configName: "core",
  network: Network.create(makeFetchQuery("/api/console/v1/graphql")),
  store: new Store(new RecordSource(), {
    queryCacheExpirationTime: 1 * 60 * 1000,
    gcReleaseBufferSize: 20,
  }),
});

export const iamEnvironment = new Environment({
  configName: "iam",
  network: Network.create(makeFetchQuery("/api/connect/v1/graphql")),
  store: new Store(new RecordSource(), {
    queryCacheExpirationTime: 1 * 60 * 1000,
    gcReleaseBufferSize: 20,
  }),
});
