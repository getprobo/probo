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

import type { RouteObject } from "react-router";
import { describe, expect, it } from "vitest";

import { routePattern } from "./route";

const routes = [
  {
    path: "organizations/:organizationId",
    children: [
      { index: true },
      { path: "risks/:riskId" },
      {
        children: [
          { path: "tasks/:taskId" },
        ],
      },
      { path: "*" },
    ],
  },
] satisfies RouteObject[];

const portalRoutes = [
  {
    path: ":employeePortalId",
    children: [
      { path: "documents/:documentId" },
    ],
  },
] satisfies RouteObject[];

describe("routePattern", () => {
  it("joins nested param paths", () => {
    expect(
      routePattern(routes, "/organizations/abc/risks/def"),
    ).toBe("/organizations/:organizationId/risks/:riskId");
  });

  it("drops the index child and keeps the parent pattern", () => {
    expect(routePattern(routes, "/organizations/abc")).toBe(
      "/organizations/:organizationId",
    );
  });

  it("skips a pathless layout", () => {
    expect(
      routePattern(routes, "/organizations/abc/tasks/task-1"),
    ).toBe("/organizations/:organizationId/tasks/:taskId");
  });

  it("keeps a splat segment", () => {
    expect(
      routePattern(routes, "/organizations/abc/unknown/extra"),
    ).toBe("/organizations/:organizationId/*");
  });

  it("matches under a basename", () => {
    expect(
      routePattern(
        portalRoutes,
        "/employee-portal/portal-1/documents/doc-1",
        "/employee-portal",
      ),
    ).toBe("/:employeePortalId/documents/:documentId");
  });

  it("returns the pathname when nothing matches", () => {
    expect(routePattern(routes, "/missing")).toBe("/missing");
  });
});
