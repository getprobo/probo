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

import { BuildingsIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { OrganizationStep_profile$key } from "#/__generated__/iam/OrganizationStep_profile.graphql";
import { RegisterDeviceCard } from "#/pages/devices/_components/RegisterDeviceCard";

const organizationStepFragment = graphql`
  fragment OrganizationStep_profile on Profile @relay(plural: true) {
    organization @required(action: THROW) {
      id
      name
    }
  }
`;

export interface OrganizationStepProps {
  profileKeys: OrganizationStep_profile$key;
  selectedOrganizationId: string | null;
  onChange: (organizationId: string) => void;
  onContinue: () => void;
}

export function OrganizationStep({
  profileKeys,
  selectedOrganizationId,
  onChange,
  onContinue,
}: OrganizationStepProps) {
  const { t } = useTranslation("enroll");
  const profiles = useFragment(organizationStepFragment, profileKeys);
  const namesById = new Map(
    profiles.map(profile => [profile.organization.id, profile.organization.name]),
  );

  return (
    <RegisterDeviceCard
      icon={<BuildingsIcon />}
      title={t("organization.title")}
      description={t("organization.description")}
      action={(
        <Button
          size={3}
          variant="solid"
          color="neutral"
          highContrast
          disabled={selectedOrganizationId === null}
          onClick={onContinue}
        >
          {t("organization.continue")}
        </Button>
      )}
    >
      <div className="w-full max-w-80">
        <Select
          value={selectedOrganizationId}
          onValueChange={(value) => {
            if (value != null) {
              onChange(value);
            }
          }}
        >
          <SelectTrigger
            size={2}
            placeholder={t("organization.placeholder")}
            className="w-full"
          >
            {(value: string | null) => (value === null ? null : namesById.get(value) ?? null)}
          </SelectTrigger>
          <SelectPopup>
            {profiles.map(profile => (
              <SelectItem key={profile.organization.id} value={profile.organization.id}>
                {profile.organization.name}
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </div>
    </RegisterDeviceCard>
  );
}
