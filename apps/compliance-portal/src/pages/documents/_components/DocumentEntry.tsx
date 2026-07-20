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

import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { Link as RouterLink } from "react-router";

import { DocumentAccessAction } from "./DocumentAccessAction";

interface DocumentEntryProps {
  // Primary line (document title, file name, or framework name).
  title: ReactNode;
  // Accent sub-label (document type, file category, or report file name).
  meta: ReactNode;
  // Whether the viewer may open the entry (public or granted access).
  isAuthorized: boolean;
  // Whether an access request is already pending for the entry.
  requested: boolean;
  // Route to the document viewer, used when authorized.
  viewHref: string;
  // Requests access for this entry (gated behind sign-in when needed).
  onGetAccess: () => void;
  // Whether the access request is in flight.
  isRequesting: boolean;
}

// Presentational row shared by the document / file / report list items: a title
// with accent metadata and the trailing access action. On small screens the
// whole row is the hit target (the trailing icon is a status affordance only).
export function DocumentEntry({
  title,
  meta,
  isAuthorized,
  requested,
  viewHref,
  onGetAccess,
  isRequesting,
}: DocumentEntryProps) {
  const { t } = useTranslation("documents");

  const mobileHitLabel = requested
    ? null
    : isAuthorized
      ? t("actions.view")
      : t("actions.getAccess");

  return (
    <ListItem
      className={[
        "relative",
        mobileHitLabel != null ? "max-sm:cursor-pointer max-sm:hover:bg-sand-2" : "",
      ].filter(Boolean).join(" ")}
    >
      <ListItemContent>
        <Text size={2} weight="medium" color="neutral" highContrast className="truncate">
          {title}
        </Text>
        <Text size={1} color="gold" className="truncate">
          {meta}
        </Text>
      </ListItemContent>

      {/* Desktop: labeled interactive control. */}
      <div className="max-sm:hidden">
        <DocumentAccessAction
          isAuthorized={isAuthorized}
          requested={requested}
          viewHref={viewHref}
          onGetAccess={onGetAccess}
          isRequesting={isRequesting}
        />
      </div>

      {/* Mobile: status icon only; the row overlay handles activation.
          Pending (`requested`) rows keep the icon in the a11y tree as status. */}
      <div
        className="hidden shrink-0 max-sm:block"
        aria-hidden={mobileHitLabel != null ? true : undefined}
      >
        <DocumentAccessAction
          isAuthorized={isAuthorized}
          requested={requested}
          viewHref={viewHref}
          onGetAccess={onGetAccess}
          isRequesting={isRequesting}
          interactive={false}
        />
      </div>

      {/* Sit above the row content so title / icon areas are part of the hit target. */}
      {mobileHitLabel != null && (
        isAuthorized
          ? (
              <RouterLink
                to={viewHref}
                className="absolute inset-0 z-1 hidden max-sm:block"
                aria-label={mobileHitLabel}
              />
            )
          : (
              <button
                type="button"
                className="absolute inset-0 z-1 hidden max-sm:block"
                aria-label={mobileHitLabel}
                disabled={isRequesting}
                onClick={onGetAccess}
              />
            )
      )}
    </ListItem>
  );
}
