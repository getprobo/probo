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

export function visitorDisplayName(fullName: string, emailAddress: string): string {
  const name = fullName.trim();
  if (name !== "") {
    return name;
  }

  return emailAddress;
}

export type VisitorAccessState = "ACTIVE" | "DEACTIVATED";

export type VisitorVisitStatus = "deactivated" | "invited" | "visited";

export function visitorVisitStatusTone(
  status: VisitorVisitStatus,
): "red" | "amber" | "faint" {
  switch (status) {
    case "deactivated":
      return "red";
    case "invited":
      return "amber";
    case "visited":
      return "faint";
  }
}

export function visitorVisitStatus(
  state: VisitorAccessState,
  authenticatedAt: string | null | undefined,
): VisitorVisitStatus {
  if (state === "DEACTIVATED") {
    return "deactivated";
  }

  if (authenticatedAt == null) {
    return "invited";
  }

  return "visited";
}
