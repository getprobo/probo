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
import type { ComponentProps, ReactNode } from "react";

import { tonedCard, type TonedCardTone } from "../_lib/tonedCard";

export type CompliancePortalTonedCardProps = ComponentProps<"div"> & {
  tone?: TonedCardTone;
  icon: ReactNode;
  lead?: ReactNode;
  control?: ReactNode;
};

// Status-shaped shell: toned frame, header icon + optional lead + control,
// body children. Domain cards keep custom JSX and call `tonedCard` directly.
export function CompliancePortalTonedCard(props: CompliancePortalTonedCardProps) {
  const { tone = "sand", icon, lead, control, className, children, ...rest } = props;
  const { frame, header, wash, fade, icon: iconSlot, lead: leadSlot, control: controlSlot, body }
    = tonedCard({ tone });

  return (
    <Card variant="ghost" size={2} padding="none" className={frame({ className })} {...rest}>
      <div className={header()}>
        <div className={wash()} />
        <div className={fade()} />
        <div className={iconSlot()}>{icon}</div>
        {lead != null && <div className={leadSlot()}>{lead}</div>}
        {control != null && <div className={controlSlot()}>{control}</div>}
      </div>
      <div className={body()}>{children}</div>
    </Card>
  );
}
