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

export function getActionLabel(
  action: string,
  __: (s: string) => string,
): string {
  switch (action) {
    case "ACCEPT_ALL":
      return __("Accept All");
    case "REJECT_ALL":
      return __("Reject All");
    case "CUSTOMIZE":
      return __("Customize");
    case "GPC":
      return __("GPC");
    default:
      return action;
  }
}

export function getActionVariant(
  action: string,
): "success" | "danger" | "warning" | "neutral" {
  switch (action) {
    case "ACCEPT_ALL":
      return "success";
    case "REJECT_ALL":
      return "danger";
    case "CUSTOMIZE":
      return "warning";
    case "GPC":
      return "neutral";
    default:
      return "neutral";
  }
}

export function formatAnonymizedIp(ip: string): string {
  if (ip.includes(".")) {
    return ip.replace(/\.0$/, ".*");
  }
  if (ip.endsWith("::")) {
    return ip + "*";
  }
  return ip;
}
