// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { Google } from "@probo/ui/src/Atoms/ThirdParties/Google";
import { Microsoft } from "@probo/ui/src/Atoms/ThirdParties/Microsoft";
import { Button } from "@probo/ui/src/v2/Button/Button";
import type { ComponentProps } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { OIDCButton_provider$key } from "./__generated__/OIDCButton_provider.graphql";

const providerFragment = graphql`
  fragment OIDCButton_provider on OIDCProviderInfo {
    name
    loginURL
  }
`;

const providerIcons: Record<string, (props: ComponentProps<"svg">) => React.ReactNode> = {
  google: Google,
  microsoft: Microsoft,
};

interface OIDCButtonProps {
  providerKey: OIDCButton_provider$key;
  // Absolute URL to return to after the provider completes authentication.
  continueTo: string;
}

// Redirects the whole window to the provider's hosted login, carrying the
// `continue` target so the portal resumes the pending flow on return.
export function OIDCButton({ providerKey, continueTo }: OIDCButtonProps) {
  const { t } = useTranslation();
  const provider = useFragment(providerFragment, providerKey);
  const Icon = providerIcons[provider.name];
  const label = t("auth.signIn.withProvider", {
    provider: provider.name.charAt(0).toUpperCase() + provider.name.slice(1),
  });

  return (
    <Button
      type="button"
      variant="soft"
      color="neutral"
      highContrast
      className="w-full"
      iconStart={Icon ? <Icon className="size-4" /> : undefined}
      onClick={() => {
        const loginURL = new URL(provider.loginURL, window.location.origin);
        loginURL.searchParams.set("continue", continueTo);
        window.location.href = loginURL.toString();
      }}
    >
      {label}
    </Button>
  );
}
