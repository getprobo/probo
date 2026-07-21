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

import type { measureStates } from "@probo/helpers";
import type { ComponentProps } from "react";
import { useTranslation } from "react-i18next";

import { Badge } from "../../Atoms/Badge/Badge";

type Props = {
  state: (typeof measureStates)[number];
};

const stateToVariant: Record<
  Props["state"],
  ComponentProps<typeof Badge>["variant"]
> = {
  IMPLEMENTED: "success",
  IN_PROGRESS: "info",
  NOT_APPLICABLE: "neutral",
  NOT_STARTED: "neutral",
  UNKNOWN: "neutral",
  NOT_IMPLEMENTED: "danger",
};

export function MeasureBadge({ state }: Props) {
  const { t } = useTranslation();
  return (
    <Badge variant={stateToVariant[state]}>
      {t(`ui.measureState.${state.toLowerCase()}`)}
    </Badge>
  );
}
