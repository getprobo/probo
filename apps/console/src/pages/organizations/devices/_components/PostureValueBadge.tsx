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

import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { PostureValueBadge_postureFragment$key } from "#/__generated__/core/PostureValueBadge_postureFragment.graphql";

import { postureValueLabel, postureValueVariant } from "../_lib/deviceDisplay";

const postureFragment = graphql`
  fragment PostureValueBadge_postureFragment on DevicePosture {
    checkKey
    status
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
  const { t } = useTranslation("organizations/devices");
  const posture = useFragment(postureFragment, postureFragmentRef);
  const { kind, text, number } = posture.value;
  const label = postureValueLabel(t, { kind, text, number });

  if (posture.status === "NOT_APPLICABLE") {
    return <Text size={2}>{label}</Text>;
  }

  if (kind === "UNKNOWN") {
    return (
      <Tooltip>
        <TooltipTrigger
          render={(
            <span tabIndex={0} className="inline-flex">
              <Badge variant="soft" color="amber">{label}</Badge>
            </span>
          )}
        />
        <TooltipPopup className="max-w-64">
          {t("devices.postures.unknownHint")}
        </TooltipPopup>
      </Tooltip>
    );
  }

  const color = postureColor(postureValueVariant(kind, posture.checkKey));

  if (!color) {
    return <Text size={2}>{label}</Text>;
  }

  return <Badge variant="soft" color={color}>{label}</Badge>;
}

function postureColor(
  variant: ReturnType<typeof postureValueVariant>,
): "green" | "red" | "sky" | undefined {
  switch (variant) {
    case "success":
      return "green";
    case "danger":
      return "red";
    case "info":
      return "sky";
    default:
      return undefined;
  }
}
