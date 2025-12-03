import { Environment, Network, RecordSource, Store } from "relay-runtime";
import { makeFetchQuery } from "@probo/relay";

const source = new RecordSource();
const store = new Store(source, {
  queryCacheExpirationTime: 1 * 60 * 1000,
  gcReleaseBufferSize: 20,
});

export const consoleEnvironment = new Environment({
  network: Network.create(makeFetchQuery("/api/console/v1/query")),
  store,
});

export const connectEnvironment = new Environment({
  network: Network.create(makeFetchQuery("/api/connect/v1/query")),
  store,
});
