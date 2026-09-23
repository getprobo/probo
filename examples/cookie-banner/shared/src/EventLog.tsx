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

import type { EventEntry } from "./useEventLog";

interface EventLogProps {
  events: EventEntry[];
}

export function EventLog({ events }: EventLogProps) {
  return (
    <section>
      <h2>Events ({events.length})</h2>
      {events.length === 0 ? (
        <p style={{ color: "#999" }}>No events yet.</p>
      ) : (
        events.map((ev) => (
          <div
            key={ev.id}
            style={{
              marginBottom: 8,
              border: "1px solid #ddd",
              padding: 8,
              background: "#f5f5f5",
            }}
          >
            <div style={{ fontWeight: "bold", marginBottom: 4 }}>
              {ev.type}{" "}
              <span style={{ fontWeight: "normal", color: "#999" }}>
                {ev.time}
              </span>
            </div>
            <pre style={{ margin: 0, overflow: "auto" }}>
              {JSON.stringify(ev.detail, null, 2)}
            </pre>
          </div>
        ))
      )}
    </section>
  );
}
