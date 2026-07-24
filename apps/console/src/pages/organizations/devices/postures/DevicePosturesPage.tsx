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

import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { DevicePosturesPageQuery } from "#/__generated__/core/DevicePosturesPageQuery.graphql";

import { DevicePostureList } from "./_components/DevicePostureList";

export const devicePosturesPageQuery = graphql`
  query DevicePosturesPageQuery($deviceId: ID!) {
    device: node(id: $deviceId) @required(action: THROW) {
      __typename
      ... on Device {
        ...DevicePostureList_deviceFragment
      }
    }
  }
`;

interface DevicePosturesPageProps {
  queryRef: PreloadedQuery<DevicePosturesPageQuery>;
}

export function DevicePosturesPage({ queryRef }: DevicePosturesPageProps) {
  const { device } = usePreloadedQuery<DevicePosturesPageQuery>(
    devicePosturesPageQuery,
    queryRef,
  );
  if (device.__typename !== "Device") {
    throw new Error("invalid type for device node");
  }

  return <DevicePostureList deviceFragmentRef={device} />;
}
