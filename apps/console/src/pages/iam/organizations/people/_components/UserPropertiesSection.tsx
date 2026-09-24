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

import { getRole, peopleRoles } from "@probo/helpers";
import { dateFormat } from "@probo/i18n";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { UserPropertiesSection_profile$key } from "#/__generated__/iam/UserPropertiesSection_profile.graphql";

import { userUpdateInput, useUpdateUser } from "../_lib/useUpdateUser";
import { isUsersListKind } from "../_lib/useUsersListFilters";
import { userPropertiesSection } from "../variants";

import { UserEmailsField } from "./UserEmailsField";

const fragment = graphql`
  fragment UserPropertiesSection_profile on Profile {
    id
    fullName
    kind
    position
    additionalEmailAddresses
    source
    contract {
      start
      end
    }
    canUpdate: permission(action: "iam:membership-profile:update")
  }
`;

interface UserPropertiesSectionProps {
  profileKey: UserPropertiesSection_profile$key;
}

export function UserPropertiesSection({ profileKey }: UserPropertiesSectionProps) {
  const { t, i18n } = useTranslation();
  const profile = useFragment(fragment, profileKey);
  const [updateUser, isUpdating] = useUpdateUser();
  const { root } = userPropertiesSection();
  const scimManaged = profile.source === "SCIM";
  const canEditIdentity = profile.canUpdate && !scimManaged;
  const canEditContract = profile.canUpdate;
  const empty = t("userPage.empty");
  const kindOptions = profile.kind != null && !isUsersListKind(profile.kind)
    ? [...peopleRoles, profile.kind]
    : peopleRoles;

  const [fullName, setFullName] = useState(profile.fullName);
  const [position, setPosition] = useState(profile.position ?? "");
  const [emails, setEmails] = useState<string[]>([...profile.additionalEmailAddresses]);
  const [contractStart, setContractStart] = useState(profile.contract?.start ?? "");
  const [contractEnd, setContractEnd] = useState(profile.contract?.end ?? "");

  useEffect(() => {
    setFullName(profile.fullName);
    setPosition(profile.position ?? "");
    setEmails([...profile.additionalEmailAddresses]);
    setContractStart(profile.contract?.start ?? "");
    setContractEnd(profile.contract?.end ?? "");
  }, [
    profile.fullName,
    profile.position,
    profile.additionalEmailAddresses,
    profile.contract,
  ]);

  function save(patch: Parameters<typeof userUpdateInput>[1]) {
    void updateUser({
      variables: {
        input: userUpdateInput({
          id: profile.id,
          fullName: profile.fullName,
          kind: profile.kind,
          position: profile.position,
          additionalEmailAddresses: profile.additionalEmailAddresses,
          contract: profile.contract,
        }, patch),
      },
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  function saveName() {
    const next = fullName.trim();
    if (next === "" || next === profile.fullName) {
      setFullName(profile.fullName);
      return;
    }
    save({ fullName: next });
  }

  function savePosition() {
    const next = position.trim();
    const current = profile.position ?? "";
    if (next === current) {
      return;
    }
    save({ position: next === "" ? null : next });
  }

  function saveEmails() {
    const next = emails.map(email => email.trim()).filter(email => email !== "");
    const current = [...profile.additionalEmailAddresses];
    if (next.length === current.length && next.every((email, index) => email === current[index])) {
      return;
    }
    save({ additionalEmailAddresses: next });
  }

  function saveContract(nextStart: string, nextEnd: string) {
    const start = toInputDate(nextStart);
    const end = toInputDate(nextEnd);
    const currentStart = toInputDate(profile.contract?.start);
    const currentEnd = toInputDate(profile.contract?.end);
    if (start === currentStart && end === currentEnd) {
      return;
    }
    save({
      contractStart: start === "" ? null : start,
      contractEnd: end === "" ? null : end,
    });
  }

  return (
    <Card variant="soft" size={2}>
      <div className={root()}>
        <PropertyRow label={t("userForm.fields.fullName")}>
          {canEditIdentity
            ? (
                <TextField
                  size={1}
                  value={fullName}
                  disabled={isUpdating}
                  required
                  aria-label={t("userForm.fields.fullName")}
                  onValueChange={setFullName}
                  onBlur={saveName}
                />
              )
            : (
                <Text size={2}>{profile.fullName}</Text>
              )}
        </PropertyRow>
        <PropertyRow label={t("userForm.fields.type")}>
          {canEditIdentity
            ? (
                <Select
                  value={profile.kind}
                  disabled={isUpdating}
                  onValueChange={(kind) => {
                    if ((kind ?? null) === (profile.kind ?? null)) {
                      return;
                    }
                    save({ kind });
                  }}
                >
                  <SelectTrigger size={1} aria-label={t("userForm.fields.type")}>
                    {(kind: string | null) => (
                      kind != null ? getRole(t, kind) : empty
                    )}
                  </SelectTrigger>
                  <SelectPopup>
                    <SelectItem value={null}>{empty}</SelectItem>
                    {kindOptions.map(kind => (
                      <SelectItem key={kind} value={kind}>
                        {isUsersListKind(kind) ? t(`userForm.kinds.${kind}`) : kind}
                      </SelectItem>
                    ))}
                  </SelectPopup>
                </Select>
              )
            : (
                <Text size={2} color={profile.kind == null ? "faint" : undefined}>
                  {profile.kind != null ? getRole(t, profile.kind) : empty}
                </Text>
              )}
        </PropertyRow>
        <PropertyRow label={t("userForm.fields.position")}>
          {canEditIdentity
            ? (
                <TextField
                  size={1}
                  value={position}
                  disabled={isUpdating}
                  placeholder={t("userForm.fields.positionPlaceholder")}
                  aria-label={t("userForm.fields.position")}
                  onValueChange={setPosition}
                  onBlur={savePosition}
                />
              )
            : (
                <Text size={2} color={profile.position == null ? "faint" : undefined}>
                  {profile.position ?? empty}
                </Text>
              )}
        </PropertyRow>
        <PropertyRow label={t("userForm.fields.additionalEmails")}>
          <UserEmailsField
            value={emails}
            disabled={isUpdating || !canEditIdentity}
            readOnly={!canEditIdentity}
            onValueChange={setEmails}
            onBlur={saveEmails}
          />
        </PropertyRow>
        <PropertyRow label={t("userForm.fields.contractStartDate")}>
          {canEditContract
            ? (
                <TextField
                  size={1}
                  type="date"
                  value={toInputDate(contractStart)}
                  disabled={isUpdating}
                  aria-label={t("userForm.fields.contractStartDate")}
                  onChange={(event) => {
                    const next = event.currentTarget.value;
                    setContractStart(next);
                    saveContract(next, contractEnd);
                  }}
                />
              )
            : (
                <ContractDateValue
                  value={profile.contract?.start}
                  language={i18n.language}
                  empty={empty}
                />
              )}
        </PropertyRow>
        <PropertyRow label={t("userForm.fields.contractEndDate")}>
          {canEditContract
            ? (
                <TextField
                  size={1}
                  type="date"
                  value={toInputDate(contractEnd)}
                  disabled={isUpdating}
                  aria-label={t("userForm.fields.contractEndDate")}
                  onChange={(event) => {
                    const next = event.currentTarget.value;
                    setContractEnd(next);
                    saveContract(contractStart, next);
                  }}
                />
              )
            : (
                <ContractDateValue
                  value={profile.contract?.end}
                  language={i18n.language}
                  empty={empty}
                />
              )}
        </PropertyRow>
      </div>
    </Card>
  );
}

function toInputDate(value: string | null | undefined) {
  if (value == null || value === "") {
    return "";
  }
  return value.split("T")[0] ?? "";
}

function PropertyRow({ label, children }: { label: string; children: ReactNode }) {
  const { row } = userPropertiesSection();

  return (
    <div className={row()}>
      <Text size={2} color="faint">{label}</Text>
      {children}
    </div>
  );
}

function ContractDateValue({
  value,
  language,
  empty,
}: {
  value: string | null | undefined;
  language: string;
  empty: string;
}) {
  if (value == null) {
    return <Text size={2} color="faint">{empty}</Text>;
  }

  return (
    <Text size={2}>
      <time dateTime={toInputDate(value)}>{dateFormat(language, value)}</time>
    </Text>
  );
}
