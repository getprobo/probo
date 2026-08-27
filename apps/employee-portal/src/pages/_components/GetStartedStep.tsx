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

import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";

import { getStartedStep } from "./variants";

export interface GetStartedStepProps {
  index: number;
  title: string;
  description: string;
  actionLabel: string;
  to: string;
  tone: "current" | "upcoming";
}

export function GetStartedStep({
  index,
  title,
  description,
  actionLabel,
  to,
  tone,
}: GetStartedStepProps) {
  const slots = getStartedStep({ tone });

  return (
    <div className={slots.row()}>
      <div className={slots.lead()}>
        <span className={slots.index()}>{index}</span>
        <div className={slots.copy()}>
          <Heading level={3} size={2} weight="medium" highContrast>
            {title}
          </Heading>
          <Text size={1} color="current" className={slots.description()}>
            {description}
          </Text>
        </div>
      </div>
      <ButtonLink
        to={to}
        size={2}
        variant={tone === "current" ? "solid" : "soft"}
        color="neutral"
        highContrast={tone === "current"}
      >
        {actionLabel}
      </ButtonLink>
    </div>
  );
}
