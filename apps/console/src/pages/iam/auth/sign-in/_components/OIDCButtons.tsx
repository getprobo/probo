import { useTranslate } from "@probo/i18n";
import { Button, Google, Microsoft } from "@probo/ui";
import type { ComponentProps } from "react";

const providerIcons: Record<
  string,
  (props: ComponentProps<"svg">) => React.ReactNode
> = {
  google: Google,
  microsoft: Microsoft,
};

export function OIDCButtons({
  providers,
  safeContinueUrl,
}: {
  providers: ReadonlyArray<{ readonly name: string; readonly loginURL: string }>;
  safeContinueUrl: URL;
}) {
  const { __ } = useTranslate();

  if (providers.length === 0) {
    return null;
  }

  return (
    <>
      {providers.map((provider) => {
        const Icon = providerIcons[provider.name];
        return (
          <Button
            key={provider.name}
            variant="secondary"
            className="w-full h-10"
            onClick={() => {
              window.location.href
                = provider.loginURL
                  + "?continue="
                  + encodeURIComponent(safeContinueUrl.toString());
            }}
          >
            <span className="flex items-center gap-2">
              {Icon && <Icon width={18} height={18} />}
              {__(`Sign in with ${provider.name.charAt(0).toUpperCase() + provider.name.slice(1)}`)}
            </span>
          </Button>
        );
      })}
    </>
  );
}
