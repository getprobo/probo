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

import { Badge } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { PostureValueBadge_postureFragment$key } from "#/__generated__/core/PostureValueBadge_postureFragment.graphql";

import { postureValueLabel, postureValueVariant } from "../_lib/deviceDisplay";

const postureFragment = graphql`
  fragment PostureValueBadge_postureFragment on DevicePosture {
    checkKey
    value {
      kind
      text
      number
    }
  }
`;

interface PostureValueBadgeProps {
  postureFragmentRef: PostureValueBadge_postureFragment$key;
}

export function PostureValueBadge({
  postureFragmentRef,
}: PostureValueBadgeProps) {
  const { t } = useTranslation();
  const posture = useFragment(postureFragment, postureFragmentRef);

  const { kind, text, number } = posture.value;
  const label = postureValueLabel(t, { kind, text, number });
  const variant = postureValueVariant(kind, posture.checkKey);

  if (!variant) {
    return label;
  }

  return <Badge variant={variant}>{label}</Badge>;
}
