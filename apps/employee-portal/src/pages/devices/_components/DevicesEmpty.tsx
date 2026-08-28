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

import { LaptopIcon } from "@phosphor-icons/react";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { devicesEmpty } from "./variants";

export interface DevicesEmptyProps {
  action?: ReactNode;
}

export function DevicesEmpty({ action }: DevicesEmptyProps) {
  const { t } = useTranslation("devices");
  const slots = devicesEmpty();

  return (
    <div className={slots.frame()}>
      <div className={slots.wash()} />
      <div className={slots.content()}>
        <div className={slots.copy()}>
          <span className={slots.icon()}>
            <LaptopIcon />
          </span>
          <Heading level={2} size={6} weight="medium" highContrast align="center">
            {t("empty.title")}
          </Heading>
          <Text size={2} color="current" align="center" className={slots.description()}>
            {t("empty.description")}
          </Text>
        </div>
        {action}
      </div>
    </div>
  );
}
