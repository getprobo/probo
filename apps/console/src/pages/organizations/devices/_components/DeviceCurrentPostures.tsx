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

import { ShieldCheckIcon } from "@phosphor-icons/react";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DeviceCurrentPostures_deviceFragment$key } from "#/__generated__/core/DeviceCurrentPostures_deviceFragment.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";

import { postureValueVariant } from "../_lib/deviceDisplay";
import { getPostureCheckLabel } from "../_lib/getPostureCheckLabel";

import { PostureValueBadge } from "./PostureValueBadge";
import { deviceCardFields } from "./variants";

const deviceFragment = graphql`
  fragment DeviceCurrentPostures_deviceFragment on Device {
    latestPostures {
      id
      checkKey
      value {
        kind
      }
      ...PostureValueBadge_postureFragment
    }
  }
`;

interface DeviceCurrentPosturesProps {
  deviceFragmentRef: DeviceCurrentPostures_deviceFragment$key;
}

export function DeviceCurrentPostures({
  deviceFragmentRef,
}: DeviceCurrentPosturesProps) {
  const { t } = useTranslation("organizations/devices");
  const device = useFragment(deviceFragment, deviceFragmentRef);
  const { root, property } = deviceCardFields();
  const tone = device.latestPostures.length === 0
    ? "sand"
    : device.latestPostures.some(
      posture => postureValueVariant(posture.value.kind, posture.checkKey) === "danger",
    )
      ? "amber"
      : "green";

  return (
    <TonedCard
      tone={tone}
      fill={false}
      icon={<ShieldCheckIcon weight="duotone" />}
      lead={(
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("devices.postures.currentTitle")}
        </Heading>
      )}
    >
      {device.latestPostures.length === 0
        ? <Text size={2} color="faint" align="center">{t("devices.postures.empty")}</Text>
        : (
            <div className={root()}>
              {device.latestPostures.map(posture => (
                <div key={posture.id} className={property()}>
                  <Text size={1} color="faint">
                    {getPostureCheckLabel(t, posture.checkKey)}
                  </Text>
                  <PostureValueBadge postureFragmentRef={posture} />
                </div>
              ))}
            </div>
          )}
    </TonedCard>
  );
}
