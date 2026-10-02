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

import { type ReactNode } from "react";
import { graphql, useFragment } from "react-relay";

import type { OAuthConnectForm_provider$key } from "#/__generated__/core/OAuthConnectForm_provider.graphql";

import { connectOAuthProvider } from "../_lib/startConnect";

import { ConnectForm } from "./ConnectForm";

const oauthConnectFormFragment = graphql`
  fragment OAuthConnectForm_provider on ConnectorProviderInfo {
    provider
    oauth2Scopes
    ...ConnectForm_provider
  }
`;

export function OAuthConnectForm({
  organizationId,
  providerKey,
  extras,
  canSubmit = true,
  children,
}: {
  organizationId: string;
  providerKey: OAuthConnectForm_provider$key;
  extras?: Record<string, string>;
  canSubmit?: boolean;
  children?: ReactNode;
}) {
  const provider = useFragment(oauthConnectFormFragment, providerKey);

  return (
    <ConnectForm
      providerKey={provider}
      canSubmit={canSubmit}
      onSubmit={({ name }) => {
        connectOAuthProvider(organizationId, provider.provider, provider.oauth2Scopes, extras, name);
        return null;
      }}
    >
      {children}
    </ConnectForm>
  );
}
