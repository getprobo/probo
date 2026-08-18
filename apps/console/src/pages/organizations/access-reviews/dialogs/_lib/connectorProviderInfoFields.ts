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

import { graphql, readFragment } from "relay-runtime";

import type { connectorProviderInfoFields_installableProtocols$key } from "#/__generated__/core/connectorProviderInfoFields_installableProtocols.graphql";
import type { connectorProviderInfoFields_oauthConfigured$key } from "#/__generated__/core/connectorProviderInfoFields_oauthConfigured.graphql";

const PROTOCOL_OAUTH2 = "OAUTH2";

/**
 * @relayField ConnectorProviderInfo.oauthConfigured: Boolean
 * @rootFragment connectorProviderInfoFields_oauthConfigured
 */
export function oauthConfigured(
  key: connectorProviderInfoFields_oauthConfigured$key,
): boolean {
  const provider = readFragment(
    graphql`
      fragment connectorProviderInfoFields_oauthConfigured on ConnectorProviderInfo {
        configuredProtocols
      }
    `,
    key,
  );

  return provider.configuredProtocols.includes(PROTOCOL_OAUTH2);
}

/**
 * @relayField ConnectorProviderInfo.installableProtocols: [String!]
 * @rootFragment connectorProviderInfoFields_installableProtocols
 */
export function installableProtocols(
  key: connectorProviderInfoFields_installableProtocols$key,
): ReadonlyArray<string> {
  const provider = readFragment(
    graphql`
      fragment connectorProviderInfoFields_installableProtocols on ConnectorProviderInfo {
        configuredProtocols
      }
    `,
    key,
  );

  return provider.configuredProtocols.filter(
    (protocol) => protocol !== PROTOCOL_OAUTH2,
  );
}
