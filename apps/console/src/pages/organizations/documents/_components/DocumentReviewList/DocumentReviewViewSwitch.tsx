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

import { ChatCircleIcon, ClockCounterClockwiseIcon } from "@phosphor-icons/react";
import { SegmentedControl } from "@probo/ui/src/v2/SegmentedControl/SegmentedControl";
import { SegmentedControlItem } from "@probo/ui/src/v2/SegmentedControl/SegmentedControlItem";
import { useTranslation } from "react-i18next";

export type DocumentReviewView = "open" | "all";

function isDocumentReviewView(value: string): value is DocumentReviewView {
  return value === "open" || value === "all";
}

interface DocumentReviewViewSwitchProps {
  view: DocumentReviewView;
  onViewChange: (view: DocumentReviewView) => void;
}

export function DocumentReviewViewSwitch({ view, onViewChange }: DocumentReviewViewSwitchProps) {
  const { t } = useTranslation("organizations/documents");

  return (
    <SegmentedControl
      size={1}
      value={view}
      aria-label={t("details.review.views.label")}
      onValueChange={(value) => {
        if (isDocumentReviewView(value)) {
          onViewChange(value);
        }
      }}
    >
      <SegmentedControlItem value="open">
        <ChatCircleIcon />
        {t("details.review.views.open")}
      </SegmentedControlItem>
      <SegmentedControlItem value="all">
        <ClockCounterClockwiseIcon />
        {t("details.review.views.all")}
      </SegmentedControlItem>
    </SegmentedControl>
  );
}
