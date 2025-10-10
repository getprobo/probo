export function getLogoUrl(logoPath: string): string {
  const path = window.location.pathname;
  const trustMatch = path.match(/^\/trust\/([^/]+)/);

  if (!trustMatch) {
    return `/logos/${logoPath}`;
  }

  const slug = trustMatch[1];
  return `/trust/${slug}/logos/${logoPath}`;
}

export function getTrustCenterUrl(path: string): string {
  const currentPath = window.location.pathname;
  const trustMatch = currentPath.match(/^\/trust\/([^/]+)/);

  if (!trustMatch) {
    return `/${path}`;
  }

  return `../${path}`;
}
