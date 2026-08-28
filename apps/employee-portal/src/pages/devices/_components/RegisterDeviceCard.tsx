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
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";

import { registerDeviceCard } from "./variants";

export interface RegisterDeviceCardProps {
  icon: ReactNode;
  title: string;
  description: string;
  children?: ReactNode;
  action?: ReactNode;
}

export function RegisterDeviceCard({
  icon,
  title,
  description,
  children,
  action,
}: RegisterDeviceCardProps) {
  const slots = registerDeviceCard();

  return (
    <Card variant="soft" size={3} padding="none" className={slots.frame()}>
      <div className={slots.header()}>
        <span aria-hidden className={slots.icon()}>{icon}</span>
        <Heading level={2} size={6} weight="medium" highContrast>
          {title}
        </Heading>
        <Text size={2} color="neutral">
          {description}
        </Text>
      </div>
      {children != null && (
        <div className={slots.body()}>
          {children}
        </div>
      )}
      {action}
    </Card>
  );
}
