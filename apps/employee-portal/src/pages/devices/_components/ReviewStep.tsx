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

import {
  CheckIcon,
  DesktopIcon,
  IdentificationCardIcon,
  LaptopIcon,
  PulseIcon,
  ShieldCheckIcon,
} from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { RegisterDeviceCard } from "./RegisterDeviceCard";
import { reviewGrid } from "./variants";

export interface ReviewStepProps {
  onContinue: () => void;
}

export function ReviewStep({ onContinue }: ReviewStepProps) {
  const { t } = useTranslation("devices");
  const slots = reviewGrid();

  return (
    <RegisterDeviceCard
      icon={<ShieldCheckIcon />}
      title={t("review.title")}
      description={t("review.description")}
      action={(
        <Button size={3} variant="solid" color="neutral" highContrast onClick={onContinue}>
          {t("review.understand")}
        </Button>
      )}
    >
      <div className={slots.root()}>
        <ReviewCell
          icon={<LaptopIcon />}
          title={t("review.thisDevice.title")}
          items={[
            t("review.thisDevice.hostname"),
            t("review.thisDevice.platform"),
            t("review.thisDevice.osVersion"),
          ]}
        />
        <ReviewCell
          icon={<IdentificationCardIcon />}
          title={t("review.identity.title")}
          items={[
            t("review.identity.hardwareUuid"),
            t("review.identity.hostname"),
            t("review.identity.serialNumber"),
          ]}
        />
        <ReviewCell
          icon={<DesktopIcon />}
          title={t("review.system.title")}
          items={[
            t("review.system.platform"),
            t("review.system.osVersion"),
            t("review.system.agentVersion"),
          ]}
        />
        <ReviewCell
          icon={<PulseIcon />}
          title={t("review.activity.title")}
          items={[
            t("review.activity.enrollmentTime"),
            t("review.activity.heartbeats"),
            t("review.activity.posture"),
          ]}
        />
      </div>
    </RegisterDeviceCard>
  );
}

function ReviewCell({
  icon,
  title,
  items,
}: {
  icon: ReactNode;
  title: string;
  items: string[];
}) {
  const slots = reviewGrid();

  return (
    <div className={slots.cell()}>
      <div className={slots.heading()}>
        <span className={slots.headingIcon()}>{icon}</span>
        <Text size={2} weight="medium" highContrast>
          {title}
        </Text>
      </div>
      <ul className={slots.items()}>
        {items.map(item => (
          <li key={item} className={slots.item()}>
            <CheckIcon className={slots.itemIcon()} />
            <Text size={2} highContrast>
              {item}
            </Text>
          </li>
        ))}
      </ul>
    </div>
  );
}
