import "react";

type CE<T = object> = React.DetailedHTMLProps<
  React.HTMLAttributes<HTMLElement> & T,
  HTMLElement
>;

declare module "react" {
  namespace JSX {
    interface IntrinsicElements {
      "probo-cookie-banner-root": CE<{
        "banner-id"?: string;
        "base-url"?: string;
        lang?: string;
      }>;
      "probo-banner": CE;
      "probo-preference-panel": CE;
      "probo-category-list": CE;
      "probo-category-toggle": CE;
      "probo-cookie-list": CE;
      "probo-accept-button": CE;
      "probo-reject-button": CE;
      "probo-customize-button": CE;
      "probo-save-button": CE;
      "probo-settings-button": CE<{ position?: string }>;
      "probo-cookie-banner": CE<{
        "banner-id"?: string;
        "base-url"?: string;
        position?: string;
        "reopen-widget"?: string;
        lang?: string;
      }>;
    }
  }
}
