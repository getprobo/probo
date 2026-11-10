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

const THEMED_PORT = "5180";
const THEMED_TCF_PORT = "5181";

export function exampleDomain(): string {
  return (import.meta.env.PUBLIC_EXAMPLE_DOMAIN ?? "").trim().replace(/^\.+|\.+$/g, "");
}

function originFor(label: string, port: string): string | null {
  const domain = exampleDomain();
  if (domain === "") {
    return null;
  }

  const url = new URL("http://placeholder");
  url.hostname = `${label}.${domain}`;
  url.port = port;
  return url.origin;
}

export function themedExampleOrigin(): string | null {
  return originFor("shop", THEMED_PORT);
}

export function themedTcfExampleOrigin(): string | null {
  return originFor("blog", THEMED_TCF_PORT);
}

// Sets the parent-domain cookie the reader reports as pre-existing.
// Call this only from the writer. A host-only cookie never reaches the
// other host, and Domain=localhost is rejected by Chrome.
export function setParentDomainExampleCookie(): void {
  const domain = exampleDomain();
  if (domain === "" || !domain.includes(".")) {
    return;
  }

  document.cookie = `probo_example=1; Domain=${domain}; Path=/`;
}
