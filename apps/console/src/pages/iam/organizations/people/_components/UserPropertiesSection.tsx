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

import { dateFormat } from "@probo/i18n";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { DateField } from "@probo/ui/src/v2/form/DateField";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { UserPropertiesSection_profile$key } from "#/__generated__/iam/UserPropertiesSection_profile.graphql";

import { userUpdateInput, useUpdateUser } from "../_lib/useUpdateUser";
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
  const { root, intro, fields } = userPropertiesSection();
  const scimManaged = profile.source === "SCIM";
  const canEditIdentity = profile.canUpdate && !scimManaged;
  const canEditContract = profile.canUpdate;
  const empty = t("userPage.empty");

  const [position, setPosition] = useState(profile.position ?? "");
  const [emails, setEmails] = useState<string[]>([...profile.additionalEmailAddresses]);
  const [contractStart, setContractStart] = useState(profile.contract?.start ?? "");
  const [contractEnd, setContractEnd] = useState(profile.contract?.end ?? "");

  useEffect(() => {
    setPosition(profile.position ?? "");
    setEmails([...profile.additionalEmailAddresses]);
    setContractStart(profile.contract?.start ?? "");
    setContractEnd(profile.contract?.end ?? "");
  }, [
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
    <section className={root()}>
      <div className={intro()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("userPage.details.title")}
        </Heading>
        <Text size={2} color="neutral">
          {t("userPage.details.description")}
        </Text>
      </div>
      <Card variant="soft" size={2}>
        <div className={fields()}>
          <Field label={t("userForm.fields.position")}>
          {canEditIdentity
            ? (
                <TextField
                  size={2}
                  value={position}
                  disabled={isUpdating}
                  placeholder={t("userForm.fields.positionPlaceholder")}
                  onValueChange={setPosition}
                  onBlur={savePosition}
                />
              )
            : (
                <Text size={2} color={profile.position == null ? "faint" : undefined}>
                  {profile.position ?? empty}
                </Text>
              )}
          </Field>
          <Field label={t("userForm.fields.additionalEmails")}>
            <UserEmailsField
              value={emails}
              disabled={isUpdating || !canEditIdentity}
              readOnly={!canEditIdentity}
              onValueChange={setEmails}
              onBlur={saveEmails}
            />
          </Field>
          <Field label={t("userForm.fields.contractStartDate")}>
            {canEditContract
              ? (
                  <DateField
                    size={2}
                    value={toInputDate(contractStart)}
                    locale={i18n.language}
                    disabled={isUpdating}
                    nullable
                    onValueChange={(next) => {
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
          </Field>
          <Field label={t("userForm.fields.contractEndDate")}>
            {canEditContract
              ? (
                  <DateField
                    size={2}
                    value={toInputDate(contractEnd)}
                    locale={i18n.language}
                    disabled={isUpdating}
                    nullable
                    onValueChange={(next) => {
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
          </Field>
        </div>
      </Card>
    </section>
  );
}

function toInputDate(value: string | null | undefined) {
  if (value == null || value === "") {
    return "";
  }
  return value.split("T")[0] ?? "";
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
