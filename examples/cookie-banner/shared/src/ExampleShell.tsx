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

import type { ReactNode } from "react";
import { EXAMPLE_APPS, type ExampleApp } from "./apps";

interface ExampleShellProps {
  current: ExampleApp["id"];
  title: string;
  description: ReactNode;
  children: ReactNode;
}

function siblingHref(port: number): string {
  const hostname =
    typeof window === "undefined" ? "localhost" : window.location.hostname;
  const protocol =
    typeof window === "undefined" ? "http:" : window.location.protocol;
  return `${protocol}//${hostname}:${port}/`;
}

export function ExampleShell({
  current,
  title,
  description,
  children,
}: ExampleShellProps) {
  return (
    <div
      style={{
        fontFamily: "system-ui, sans-serif",
        maxWidth: 900,
        margin: "0 auto",
        padding: 24,
      }}
    >
      <nav
        style={{
          display: "flex",
          gap: 0,
          borderBottom: "2px solid #ddd",
          marginBottom: 24,
        }}
      >
        {EXAMPLE_APPS.map((app) => {
          const active = app.id === current;
          return (
            <a
              key={app.id}
              href={siblingHref(app.port)}
              style={{
                padding: "8px 16px",
                borderBottom: active
                  ? "2px solid #333"
                  : "2px solid transparent",
                color: "#111",
                textDecoration: "none",
                fontWeight: active ? "bold" : "normal",
                marginBottom: -2,
                fontSize: 14,
              }}
            >
              {app.label}
            </a>
          );
        })}
      </nav>

      <h1 style={{ marginBottom: 4 }}>{title}</h1>
      <p style={{ color: "#666", marginTop: 0, marginBottom: 24 }}>
        {description}
      </p>

      {children}
    </div>
  );
}
