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

export function formatDatetime(dateString?: string | null): string | undefined {
  if (!dateString) return undefined;
  return `${dateString}T00:00:00Z`;
}

export function toDateInput(dateString?: string | null): string {
  if (!dateString) return '';
  return dateString.split('T')[0];
}

export function todayAsDateInput(): string {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function formatDate(dateInput?: string | null): string {
  if (!dateInput) return "";

  const date = parseDate(dateInput);
  return date.toLocaleDateString();
}

export function parseDate(dateString: string): Date {
  if (dateString.includes("T")) {
    return new Date(dateString);
  }
  const parts = dateString.split("-");
  return new Date(
    parseInt(parts[0], 10),
    parts[1] ? parseInt(parts[1], 10) - 1 : 0,
    parts[2] ? parseInt(parts[2], 10) : 1,
  );
}

export function formatDuration(duration?: string | null, __?: (s: string) => string): string | null {
  if (!duration || !__) return null;

  const timeMatch = duration.match(/PT(\d+)([MH])/);
  if (timeMatch) {
    const amount = parseInt(timeMatch[1], 10) || 0;
    const unit = timeMatch[2];
    if (unit === "M") {
      return `${amount} ${amount === 1 ? __("Minute") : __("Minutes")}`;
    } else if (unit === "H") {
      return `${amount} ${amount === 1 ? __("Hour") : __("Hours")}`;
    }
  }

  const dateMatch = duration.match(/P(\d+)([DW])/);
  if (dateMatch) {
    const amount = parseInt(dateMatch[1], 10) || 0;
    const unit = dateMatch[2];
    if (unit === "W") {
      return `${amount} ${amount === 1 ? __("Week") : __("Weeks")}`;
    } else if (unit === "D") {
      const days = amount;
      if (days % 7 === 0 && days > 0) {
        const weeks = days / 7;
        return `${weeks} ${weeks === 1 ? __("Week") : __("Weeks")}`;
      }
      return `${days} ${days === 1 ? __("Day") : __("Days")}`;
    }
  }

  return null;
}
