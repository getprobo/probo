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

import { TrashIcon, WarningCircleIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { WorkspaceDangerZoneSectionFragment$key } from "#/__generated__/iam/WorkspaceDangerZoneSectionFragment.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";

import { dangerZoneSection } from "../variants";

import { DeleteWorkspaceDialog } from "./DeleteWorkspaceDialog";

const SETTINGS_NS = "iam/organizations/settings";

const fragment = graphql`
  fragment WorkspaceDangerZoneSectionFragment on Organization {
    id
    name @required(action: THROW)
    canDelete: permission(action: "iam:organization:delete")
  }
`;

interface WorkspaceDangerZoneSectionProps {
  organizationKey: WorkspaceDangerZoneSectionFragment$key;
}

export function WorkspaceDangerZoneSection({
  organizationKey,
}: WorkspaceDangerZoneSectionProps) {
  const { t } = useTranslation(SETTINGS_NS);
  const { root, intro, actions } = dangerZoneSection();
  const organization = useFragment(fragment, organizationKey);

  if (!organization.canDelete) {
    return null;
  }

  return (
    <section className={root()}>
      <div className={intro()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("dangerZone.title")}
        </Heading>
      </div>
      <TonedCard
        tone="red"
        icon={<WarningCircleIcon size={24} weight="duotone" />}
        lead={(
          <Text size={3} weight="medium" highContrast className="truncate">
            {t("dangerZone.delete.title")}
          </Text>
        )}
      >
        <Text size={2} color="neutral">
          {t("dangerZone.delete.description")}
        </Text>
        <Text size={2} color="red" weight="medium">
          {t("dangerZone.delete.warning")}
        </Text>
        <div className={actions()}>
          <DeleteWorkspaceDialog
            organizationId={organization.id}
            organizationName={organization.name}
          >
            <Button
              type="button"
              variant="solid"
              color="red"
              iconStart={<TrashIcon />}
            >
              {t("dangerZone.delete.action")}
            </Button>
          </DeleteWorkspaceDialog>
        </div>
      </TonedCard>
    </section>
  );
}
