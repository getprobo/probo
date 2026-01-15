import { matchPath } from "react-router";

export function getPathPrefix() {
  const match = matchPath(
    { path: "/trust/:id", caseSensitive: false, end: false },
    window.location.pathname,
  );

  let prefix = "";
  if (match) {
    prefix = `/trust/${match.params.id}`;
  }

  return prefix;
}
