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

import { Text } from "@probo/ui/src/v2/typography/Text";
import type { TFunction } from "i18next";
import { useTranslation } from "react-i18next";

import type { BlockLabel, SuggestionChange } from "../../_lib/reviewAnchors";
import { documentSuggestionChange } from "../../variants";

interface DocumentSuggestionChangeProps {
  change: SuggestionChange;
}

export function DocumentSuggestionChange({ change }: DocumentSuggestionChangeProps) {
  const { t } = useTranslation("organizations/documents");
  const { root, text, deleted, label } = documentSuggestionChange();

  if (change.kind === "text") {
    return (
      <div className={root()}>
        {change.deleted !== "" && (
          <Text size={2} color="red" className={deleted()}>
            {change.deleted}
          </Text>
        )}
        {change.inserted !== "" && (
          <Text size={2} color="green" className={text()}>
            {change.inserted}
          </Text>
        )}
      </div>
    );
  }

  return (
    <Text size={1} color="faint" className={label()}>
      {changeLabel(change, t)}
    </Text>
  );
}

function changeLabel(
  change: Exclude<SuggestionChange, { kind: "text" }>,
  t: TFunction<"organizations/documents">,
): string {
  switch (change.kind) {
    case "format":
      if (change.style === "link") {
        if (change.action === "remove") return t("details.review.changes.linkRemoved");
        if (change.href == null) return t("details.review.changes.formatAdded.link");
        return change.previousHref != null && change.previousHref !== change.href
          ? t("details.review.changes.linkChanged", { href: change.href })
          : t("details.review.changes.linkAdded", { href: change.href });
      }
      return change.action === "add"
        ? t(`details.review.changes.formatAdded.${change.style}`)
        : t(`details.review.changes.formatRemoved.${change.style}`);
    case "block":
      if (sameBlock(change.before, change.after)) {
        return t("details.review.changes.indentChanged");
      }
      return t("details.review.changes.blockChanged", {
        before: blockName(change.before, t),
        after: blockName(change.after, t),
      });
    case "node":
      return change.action === "insert"
        ? t(`details.review.changes.nodeAdded.${change.node}`)
        : t(`details.review.changes.nodeRemoved.${change.node}`);
  }
}

function blockName(block: BlockLabel, t: TFunction<"organizations/documents">): string {
  return t(`details.review.changes.blocks.${block.kind}`, { level: block.level ?? 1 });
}

function sameBlock(before: BlockLabel, after: BlockLabel): boolean {
  return before.kind === after.kind && before.level === after.level;
}
