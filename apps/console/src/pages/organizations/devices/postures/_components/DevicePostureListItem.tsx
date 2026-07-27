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

import { dateTimeFormat } from "@probo/i18n/date";
import { Badge, Td, Tr } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type {
  DevicePostureListItem_postureFragment$key,
} from "#/__generated__/core/DevicePostureListItem_postureFragment.graphql";

import { statusVariant } from "../../_lib/deviceDisplay";
import { getPostureCheckLabel } from "../_lib/getPostureCheckLabel";
import { getPostureStatusLabel } from "../_lib/getPostureStatusLabel";

const postureFragment = graphql`
  fragment DevicePostureListItem_postureFragment on DevicePosture {
    checkKey
    status
    observedAt
  }
`;

interface DevicePostureListItemProps {
  postureKey: DevicePostureListItem_postureFragment$key;
}

export function DevicePostureListItem({ postureKey }: DevicePostureListItemProps) {
  const { i18n, t } = useTranslation();
  const posture = useFragment(postureFragment, postureKey);

  return (
    <Tr>
      <Td>{getPostureCheckLabel(t, posture.checkKey)}</Td>
      <Td>
        <Badge variant={statusVariant(posture.status)}>
          {getPostureStatusLabel(t, posture.status)}
        </Badge>
      </Td>
      <Td className="text-end whitespace-nowrap">
        {dateTimeFormat(i18n.language, posture.observedAt)}
      </Td>
    </Tr>
  );
}
