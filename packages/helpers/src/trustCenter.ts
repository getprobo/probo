export function getTrustCenterUrl(path: string): string {
  const currentPath = window.location.pathname;
  const trustMatch = currentPath.match(/^\/trust\/([^/]+)/);

  if (!trustMatch) {
    return `/${path}`;
  }

  return `../${path}`;
}
