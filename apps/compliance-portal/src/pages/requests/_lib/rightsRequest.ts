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

import {
  ExportIcon,
  FileArrowDownIcon,
  NoteIcon,
  PencilSimpleIcon,
  ProhibitIcon,
  TrashIcon,
  WarningCircleIcon,
} from "@phosphor-icons/react";
import type { RightsRequestState, RightsRequestType } from "@probo/helpers";
import { createElement, type ReactNode } from "react";

type BadgeColor = "neutral" | "gold" | "red" | "green" | "amber" | "sky";
type BadgeVariant = "solid" | "soft" | "surface" | "outline";

// Types the portal lets a data subject submit. PORTABILITY is intentionally
// excluded pending a dedicated export/transfer flow, but still renders in the
// list (via the metadata below) when created from the console.
export const submittableRightsRequestTypes = [
  "ACCESS",
  "DELETION",
  "RECTIFICATION",
  "OBJECTION",
  "COMPLAINT",
] as const satisfies readonly RightsRequestType[];

export type SubmittableRightsRequestType = (typeof submittableRightsRequestTypes)[number];

// Icon element per type, covering all six so console-created rows (e.g.
// PORTABILITY) still render a glyph in the list. Elements are created once at
// module scope so consumers render them without creating components per render.
const rightsRequestTypeIcons: Record<RightsRequestType, ReactNode> = {
  ACCESS: createElement(FileArrowDownIcon),
  DELETION: createElement(TrashIcon),
  RECTIFICATION: createElement(PencilSimpleIcon),
  PORTABILITY: createElement(ExportIcon),
  OBJECTION: createElement(ProhibitIcon),
  COMPLAINT: createElement(WarningCircleIcon),
};

export function getRightsRequestTypeIcon(type: RightsRequestType): ReactNode {
  return rightsRequestTypeIcons[type] ?? createElement(NoteIcon);
}

// Badge treatment for the portal's status vocabulary (Pending / Processing /
// Completed / Rejected), mapped from the model's states.
export function getRightsRequestStatusBadge(
  state: RightsRequestState,
): { color: BadgeColor; variant: BadgeVariant } {
  switch (state) {
    case "TODO":
      return { color: "amber", variant: "soft" };
    case "IN_PROGRESS":
      return { color: "neutral", variant: "outline" };
    case "DONE":
      return { color: "green", variant: "soft" };
    case "REJECTED":
      return { color: "red", variant: "soft" };
    default:
      return { color: "neutral", variant: "soft" };
  }
}

// Per-type dialog form configuration. The details field's label/placeholder are
// type-specific; the deletion warning callout only shows for DELETION, and the
// name is optional for an abuse report (COMPLAINT).
export const rightsRequestFormConfig: Record<
  SubmittableRightsRequestType,
  { showDeletionWarning: boolean; nameOptional: boolean }
> = {
  ACCESS: { showDeletionWarning: false, nameOptional: false },
  DELETION: { showDeletionWarning: true, nameOptional: false },
  RECTIFICATION: { showDeletionWarning: false, nameOptional: false },
  OBJECTION: { showDeletionWarning: false, nameOptional: false },
  COMPLAINT: { showDeletionWarning: false, nameOptional: true },
};

// Human-facing reference derived from the created year and a short suffix of the
// opaque id (display-only; not a stored sequential number).
export function formatRightsRequestReference(id: string, createdAt: string): string {
  const year = new Date(createdAt).getFullYear();
  const suffix = id.replace(/[^a-zA-Z0-9]/g, "").slice(-6).toUpperCase();
  return `REQ-${year}-${suffix}`;
}
