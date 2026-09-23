import "react";

type CE<T = object> = React.DetailedHTMLProps<
  React.HTMLAttributes<HTMLElement> & T,
  HTMLElement
>;

declare module "react" {
  namespace JSX {
    interface IntrinsicElements {
      "probo-settings-link": CE;
      "probo-cookie-banner": CE<{
        "banner-id"?: string;
        "base-url"?: string;
        position?: string;
        lang?: string;
        "gcm-enabled"?: string;
      }>;
    }
  }
}
