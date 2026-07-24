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

import { useTranslate } from "@probo/i18n";
import { Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DevicePostureList_deviceFragment$key } from "#/__generated__/core/DevicePostureList_deviceFragment.graphql";

import { DevicePostureListItem } from "./DevicePostureListItem";

const deviceFragment = graphql`
  fragment DevicePostureList_deviceFragment on Device {
    latestPostures {
      id
      ...DevicePostureListItem_postureFragment
    }
  }
`;

interface DevicePostureListProps {
  deviceFragmentRef: DevicePostureList_deviceFragment$key;
}

export function DevicePostureList({ deviceFragmentRef }: DevicePostureListProps) {
  const { __ } = useTranslate();
  const device = useFragment(deviceFragment, deviceFragmentRef);

  return (
    <Table>
      <Thead>
        <Tr>
          <Th>{__("Check")}</Th>
          <Th>{__("Status")}</Th>
          <Th className="text-end">{__("Observed at")}</Th>
        </Tr>
      </Thead>
      <Tbody>
        {device.latestPostures.length === 0 && (
          <Tr>
            <Td colSpan={3} className="text-center text-txt-secondary">
              {__("No posture checks recorded")}
            </Td>
          </Tr>
        )}
        {device.latestPostures.map(posture => (
          <DevicePostureListItem key={posture.id} postureKey={posture} />
        ))}
      </Tbody>
    </Table>
  );
}
