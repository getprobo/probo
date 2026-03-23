import { useTranslate } from "@probo/i18n";
import { Button, Google, Microsoft } from "@probo/ui";
import type { ComponentProps } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { OIDCButtonsFragment$key } from "#/__generated__/iam/OIDCButtonsFragment.graphql";

const fragment = graphql`
  fragment OIDCButtonsFragment on OIDCProviderInfo {
    name
    loginURL
  }
`;

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
  providers: ReadonlyArray<OIDCButtonsFragment$key>;
  safeContinueUrl: URL;
}) {
  const { __ } = useTranslate();

  return (
    <>
      {providers.map((providerRef, index) => (
        <OIDCButton
          key={index}
          providerRef={providerRef}
          safeContinueUrl={safeContinueUrl}
          __={__}
        />
      ))}
    </>
  );
}

function OIDCButton({
  providerRef,
  safeContinueUrl,
  __,
}: {
  providerRef: OIDCButtonsFragment$key;
  safeContinueUrl: URL;
  __: (s: string) => string;
}) {
  const provider = useFragment(fragment, providerRef);
  const Icon = providerIcons[provider.name];

  return (
    <Button
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
}
