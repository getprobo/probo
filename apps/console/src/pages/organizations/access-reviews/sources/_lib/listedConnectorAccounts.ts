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

// A connector-name match lists every account. An account-name match lists only
// those accounts. No accounts, or no match, hides the connector.
export function listedConnectorAccounts<TAccount extends { name: string }>(
  accounts: readonly TAccount[],
  connector: { displayName: string; provider: string },
  normalizedSearch: string,
): readonly TAccount[] | null {
  if (accounts.length === 0) {
    return null;
  }
  if (normalizedSearch === "") {
    return accounts;
  }

  const connectorMatches = connector.displayName.toLowerCase().includes(normalizedSearch)
    || connector.provider.replaceAll("_", " ").toLowerCase().includes(normalizedSearch);
  if (connectorMatches) {
    return accounts;
  }

  const matching = accounts.filter(account =>
    account.name.toLowerCase().includes(normalizedSearch),
  );
  return matching.length > 0 ? matching : null;
}
