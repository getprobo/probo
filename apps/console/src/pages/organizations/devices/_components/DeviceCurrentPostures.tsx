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

import { Card } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DeviceCurrentPostures_deviceFragment$key } from "#/__generated__/core/DeviceCurrentPostures_deviceFragment.graphql";

import { getPostureCheckLabel } from "../_lib/getPostureCheckLabel";

import { PostureValueBadge } from "./PostureValueBadge";

const deviceFragment = graphql`
  fragment DeviceCurrentPostures_deviceFragment on Device {
    latestPostures {
      id
      checkKey
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
  const { t } = useTranslation();
  const device = useFragment(deviceFragment, deviceFragmentRef);

  return (
    <div className="space-y-4">
      <h2 className="text-base font-medium">
        {t("devices.postures.currentTitle")}
      </h2>
      <Card className="space-y-4" padded>
        {device.latestPostures.length === 0
          ? (
              <div className="text-sm text-txt-secondary">
                {t("devices.postures.empty")}
              </div>
            )
          : (
              <div className="grid grid-cols-3 gap-4">
                {device.latestPostures.map(posture => (
                  <div key={posture.id}>
                    <div className="text-xs text-txt-tertiary font-semibold mb-1">
                      {getPostureCheckLabel(t, posture.checkKey)}
                    </div>
                    <div className="text-sm text-txt-primary">
                      <PostureValueBadge postureFragmentRef={posture} />
                    </div>
                  </div>
                ))}
              </div>
            )}
      </Card>
    </div>
  );
}
