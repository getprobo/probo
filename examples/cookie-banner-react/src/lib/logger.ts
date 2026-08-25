import log from "loglevel";

function named(name: string) {
  return log.getLogger(name);
}

log.setLevel("debug");

const namedLoggers = {
  "probo:example:app": named("probo:example:app"),
  "probo:example:config": named("probo:example:config"),
  "probo:example:themed": named("probo:example:themed"),
  "probo:example:headless": named("probo:example:headless"),
  "probo:example:debug": named("probo:example:debug"),
} as const;

export const appLogger = namedLoggers["probo:example:app"];
export const configLogger = namedLoggers["probo:example:config"];
export const themedLogger = namedLoggers["probo:example:themed"];
export const headlessLogger = namedLoggers["probo:example:headless"];
export const debugLogger = namedLoggers["probo:example:debug"];

export function enableNamedLoggers(): void {
  for (const [name, logger] of Object.entries(namedLoggers)) {
    logger.setLevel("debug");
    // loglevel persists via property assignment, which the SDK does not
    // hook. Replay the same key through setItem so a live write is
    // observed as SCRIPT after the detector has wrapped Storage.
    localStorage.setItem(`loglevel:${name}`, "DEBUG");
  }
}
