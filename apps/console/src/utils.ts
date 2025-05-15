export function buildEndpoint(path: string): string {
  const host = import.meta.env.API_SERVER_HOST!;

  if (!host) {
    return path;
  }

  const formattedHost =
    host.startsWith("http://") || host.startsWith("https://")
      ? host
      : `https://${host}`;

  const url = new URL(formattedHost);

  if (path) {
    url.pathname = path.startsWith("/") ? path : `/${path}`;
  }

  return url.toString();
}
