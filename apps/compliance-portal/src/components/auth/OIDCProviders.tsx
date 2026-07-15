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

import { ButtonSkeleton } from "@probo/ui/src/v2/Button/ButtonSkeleton";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery } from "react-relay";

import type { OIDCProvidersQuery } from "./__generated__/OIDCProvidersQuery.graphql";
import { OIDCButton } from "./OIDCButton";

const oidcProvidersQuery = graphql`
  query OIDCProvidersQuery {
    oidcProviders {
      ...OIDCButton_provider
    }
  }
`;

interface OIDCProvidersProps {
  continueTo: string;
}

function Divider() {
  const { t } = useTranslation();
  return (
    <div className="flex items-center gap-4">
      <span className="h-px flex-1 bg-sand-6" />
      <Text size={1} color="faint">{t("auth.signIn.or")}</Text>
      <span className="h-px flex-1 bg-sand-6" />
    </div>
  );
}

function OIDCProvidersContent({ continueTo }: OIDCProvidersProps) {
  const data = useLazyLoadQuery<OIDCProvidersQuery>(oidcProvidersQuery, {});

  if (data.oidcProviders.length === 0) {
    return null;
  }

  return (
    <>
      <div className="flex flex-col gap-2">
        {data.oidcProviders.map((provider, index) => (
          <OIDCButton key={index} providerKey={provider} continueTo={continueTo} />
        ))}
      </div>
      <Divider />
    </>
  );
}

// SSO buttons for the sign-in dialog. Providers load lazily on open; if the
// query fails or the trust center has none, the section renders nothing so the
// email flow stays usable.
export function OIDCProviders({ continueTo }: OIDCProvidersProps) {
  return (
    <ErrorBoundary
      fallback={null}
      onError={error => console.error("Failed to load SSO providers", error)}
    >
      <Suspense
        fallback={(
          <div className="flex flex-col gap-2">
            <ButtonSkeleton className="w-full" />
            <ButtonSkeleton className="w-full" />
          </div>
        )}
      >
        <OIDCProvidersContent continueTo={continueTo} />
      </Suspense>
    </ErrorBoundary>
  );
}
