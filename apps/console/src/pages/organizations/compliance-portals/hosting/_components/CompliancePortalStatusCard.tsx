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

import { Card } from "@probo/ui/src/v2/Card/Card";
import { Switch } from "@probo/ui/src/v2/Switch/Switch";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";

import { hostingCard } from "../variants";

interface CompliancePortalStatusCardProps {
  title: string;
  description: string;
  icon: ReactNode;
  checked: boolean;
  disabled?: boolean;
  disabledHint?: string;
  onCheckedChange: (checked: boolean) => void;
}

export function CompliancePortalStatusCard({
  title,
  description,
  icon,
  checked,
  disabled,
  disabledHint,
  onCheckedChange,
}: CompliancePortalStatusCardProps) {
  const { frame, header, wash, fade, icon: iconSlot, control, body } = hostingCard({
    tone: checked ? "green" : "sand",
  });

  const toggle = (
    <Switch
      size={2}
      checked={checked}
      color="green"
      disabled={disabled}
      aria-label={title}
      onCheckedChange={onCheckedChange}
    />
  );

  return (
    <Card variant="ghost" size={2} padding="none" className={frame()}>
      <div className={header()}>
        <div className={wash()} />
        <div className={fade()} />
        <div className={iconSlot()}>{icon}</div>
        <div className={control()}>
          {disabled === true && disabledHint != null
            ? (
                <Tooltip>
                  <TooltipTrigger render={<span>{toggle}</span>} />
                  <TooltipPopup>{disabledHint}</TooltipPopup>
                </Tooltip>
              )
            : toggle}
        </div>
      </div>
      <div className={body()}>
        <Text size={3} weight="medium" highContrast>
          {title}
        </Text>
        <Text size={2} color="neutral">
          {description}
        </Text>
      </div>
    </Card>
  );
}
