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

import { Field } from "@probo/ui/src/v2/form/Field";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DeviceOwnerSelect_profile$key } from "#/__generated__/core/DeviceOwnerSelect_profile.graphql";
import { usePeople } from "#/hooks/graph/PeopleGraph";

const UNASSIGNED = "__unassigned__";

const deviceOwnerSelectProfileFragment = graphql`
  fragment DeviceOwnerSelect_profile on Profile {
    id
    fullName
  }
`;

interface DeviceOwnerSelectProps {
  organizationId: string;
  profileKey: DeviceOwnerSelect_profile$key | null;
  value: string | null;
  onValueChange: (ownerId: string | null) => void;
}

export function DeviceOwnerSelect({
  organizationId,
  profileKey,
  value,
  onValueChange,
}: DeviceOwnerSelectProps) {
  const { t } = useTranslation("organizations/devices");

  return (
    <Field label={t("devices.fields.owner")}>
      <Suspense fallback={<SelectSkeleton />}>
        <OwnerOptions
          organizationId={organizationId}
          profileKey={profileKey}
          value={value}
          onValueChange={onValueChange}
        />
      </Suspense>
    </Field>
  );
}

function OwnerOptions({
  organizationId,
  profileKey,
  value,
  onValueChange,
}: DeviceOwnerSelectProps) {
  const { t } = useTranslation("organizations/devices");
  const currentOwner = useFragment(deviceOwnerSelectProfileFragment, profileKey);
  const listed = usePeople(organizationId, { contractEnded: false });
  const people = currentOwner != null && !listed.some(person => person.id === currentOwner.id)
    ? [currentOwner, ...listed]
    : listed;
  const selected = value ?? UNASSIGNED;

  return (
    <Select
      value={selected}
      onValueChange={(next) => {
        onValueChange(next == null || next === UNASSIGNED ? null : next);
      }}
    >
      <SelectTrigger
        placeholder={t("devices.values.unassigned")}
        aria-label={t("devices.fields.owner")}
      >
        {(current: string | null) => {
          if (current == null || current === UNASSIGNED) {
            return t("devices.values.unassigned");
          }
          return people.find(person => person.id === current)?.fullName ?? current;
        }}
      </SelectTrigger>
      <SelectPopup>
        <SelectItem value={UNASSIGNED}>{t("devices.values.unassigned")}</SelectItem>
        {people.map(person => (
          <SelectItem key={person.id} value={person.id}>
            {person.fullName}
          </SelectItem>
        ))}
      </SelectPopup>
    </Select>
  );
}
