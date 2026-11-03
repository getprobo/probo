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
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";

import { sampleFamilyCounts } from "../../variants";

interface StatValueProps {
  value: string;
  label: string;
}

export function StatValue({ value, label }: StatValueProps) {
  const { rateLead, value: valueClass } = sampleFamilyCounts();

  return (
    <div className={rateLead()}>
      <Text size={4} weight="bold" highContrast className={valueClass()}>
        {value}
      </Text>
      <Text size={2}>{label}</Text>
    </div>
  );
}

interface StatCardProps {
  icon: ReactNode;
  title: string;
  value: string;
}

export function StatCard({ icon, title, value }: StatCardProps) {
  const { family, logo, value: valueClass } = sampleFamilyCounts();

  return (
    <Card
      variant="soft"
      size={3}
      padding={4}
      className="flex min-w-0 flex-col items-stretch gap-3"
    >
      <div className={family()}>
        <span className={logo()}>{icon}</span>
        <Text size={4} weight="medium">{title}</Text>
      </div>
      <Text size={4} weight="bold" highContrast className={valueClass()}>
        {value}
      </Text>
    </Card>
  );
}
