// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

type Translator = (s: string) => string;

export const snapshotTypes = [
  "RISKS",
  "VENDORS",
  "PROCESSING_ACTIVITIES",
] as const;

export function getSnapshotTypeLabel(__: Translator, type: string | null | undefined) {
  if (!type) {
    return __("Unknown");
  }

  switch (type) {
    case "RISKS":
      return __("Risks");
    case "VENDORS":
      return __("Vendors");
    case "PROCESSING_ACTIVITIES":
      return __("Processing Activities");
    default:
      return __("Unknown");
  }
}

export function getSnapshotTypeUrlPath(type?: string): string {
  switch (type) {
    case "RISKS":
      return "/risks";
    case "VENDORS":
      return "/vendors";
    case "PROCESSING_ACTIVITIES":
      return "/processing-activities";
    default:
      return "";
  }
}

export interface SnapshotableResource {
  snapshotId?: string | null | undefined;
}

export function validateSnapshotConsistency(
  resource: SnapshotableResource | null | undefined,
  urlSnapshotId?: string | null | undefined
): void {
  if (resource && resource.snapshotId !== (urlSnapshotId ?? null)) {
    throw new Error("PAGE_NOT_FOUND");
  }
}
