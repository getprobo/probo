import log from "loglevel";

function named(name: string) {
  const logger = log.getLogger(name);
  logger.setLevel("debug");
  return logger;
}

log.setLevel("debug");

export const appLogger = named("probo:example:app");
export const configLogger = named("probo:example:config");
export const themedLogger = named("probo:example:themed");
export const headlessLogger = named("probo:example:headless");
export const debugLogger = named("probo:example:debug");
