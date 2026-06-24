// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { clsx } from "clsx";

type Props = {
  level: "LOW" | "MEDIUM" | "HIGH" | "URGENT";
};

export function PriorityLevel({ level }: Props) {
  if (level === "URGENT") {
    return (
      <div className="w-max flex items-center justify-center text-txt-danger">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path
            d="M7 1.75v5.25M7 10.5h.005"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </div>
    );
  }

  const bars = level === "HIGH" ? 3 : level === "MEDIUM" ? 2 : 1;

  return (
    <div className="w-max p-[2px] flex gap-[2px] items-end">
      <div
        className={clsx(
          "h-1 w-[3px] rounded",
          bars >= 1 ? "bg-txt-secondary" : "bg-txt-quaternary",
        )}
      />
      <div
        className={clsx(
          "h-2 w-[3px] rounded",
          bars >= 2 ? "bg-txt-secondary" : "bg-txt-quaternary",
        )}
      />
      <div
        className={clsx(
          "h-3 w-[3px] rounded",
          bars >= 3 ? "bg-txt-secondary" : "bg-txt-quaternary",
        )}
      />
    </div>
  );
}
