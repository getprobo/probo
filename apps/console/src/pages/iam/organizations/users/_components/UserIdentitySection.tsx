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

import { peopleRoles } from "@probo/helpers";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { UserIdentitySection_profile$key } from "#/__generated__/iam/UserIdentitySection_profile.graphql";
import { ImageDropzone } from "#/components/ImageDropzone/ImageDropzone";

import { userUpdateInput, useUpdateUser } from "../_lib/useUpdateUser";
import { isUsersListKind } from "../_lib/useUsersListFilters";
import { userIdentitySection } from "../variants";

import { UserRoleSelect } from "./UserRoleSelect";

const fragment = graphql`
  fragment UserIdentitySection_profile on Profile {
    id
    fullName
    kind
    position
    additionalEmailAddresses
    source
    state
    avatar {
      downloadUrl
    }
    contract {
      # eslint-disable-next-line relay/unused-fields -- passed through userUpdateInput
      start
      # eslint-disable-next-line relay/unused-fields -- passed through userUpdateInput
      end
    }
    membership @required(action: THROW) {
      ...UserRoleSelect_membership
    }
    canUpdate: permission(action: "iam:membership-profile:update")
  }
`;

interface UserIdentitySectionProps {
  profileKey: UserIdentitySection_profile$key;
}

export function UserIdentitySection({ profileKey }: UserIdentitySectionProps) {
  const { t } = useTranslation();
  const profile = useFragment(fragment, profileKey);
  const [updateUser, isUpdating] = useUpdateUser();
  const isInactive = profile.state === "DEACTIVATED";
  const {
    root,
    person,
    avatar,
    dropzone,
    source,
    fields,
  } = userIdentitySection({ inactive: isInactive });
  const scimManaged = profile.source === "SCIM";
  const canEditIdentity = profile.canUpdate && !scimManaged;
  const showSource = profile.source === "SCIM" || profile.source === "SAML";
  const empty = t("userPage.empty");
  const kindOptions = profile.kind != null && !isUsersListKind(profile.kind)
    ? [...peopleRoles, profile.kind]
    : peopleRoles;

  const [fullName, setFullName] = useState(profile.fullName);
  const [nameSource, setNameSource] = useState(profile.fullName);

  if (profile.fullName !== nameSource) {
    setNameSource(profile.fullName);
    setFullName(profile.fullName);
  }

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

  function kindLabel(kind: string | null | undefined) {
    if (kind == null) {
      return empty;
    }
    return isUsersListKind(kind) ? t(`userForm.kinds.${kind}`) : kind;
  }

  return (
    <Card variant="soft" size={2}>
      <div className={root()}>
        <div className={person()}>
          <div className={avatar()}>
            <div className={dropzone()}>
              <ImageDropzone
                ratio="square"
                src={profile.avatar?.downloadUrl}
                disabled
                placeholder={t("userPage.fields.avatarPlaceholder")}
                onFile={() => {
                  // UpdateUser does not accept an avatar file yet.
                }}
                onReject={() => {
                  // Dropzone stays disabled until avatar upload is wired.
                }}
              />
              {showSource && (
                <span className={source()}>
                  <Badge variant="soft" color="neutral" size={1}>
                    {profile.source}
                  </Badge>
                </span>
              )}
            </div>
          </div>
          <div className={fields()}>
            <Field label={t("userForm.fields.fullName")}>
              {canEditIdentity
                ? (
                    <TextField
                      size={2}
                      value={fullName}
                      disabled={isUpdating}
                      required
                      onValueChange={setFullName}
                      onBlur={saveName}
                    />
                  )
                : (
                    <Text size={2} highContrast>{profile.fullName}</Text>
                  )}
            </Field>
            <Field label={t("userForm.fields.type")}>
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
                      <SelectTrigger size={2} aria-label={t("userForm.fields.type")}>
                        {(kind: string | null) => kindLabel(kind)}
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
                      {kindLabel(profile.kind)}
                    </Text>
                  )}
            </Field>
            <Field label={t("userForm.fields.role")}>
              <UserRoleSelect membershipKey={profile.membership} size={2} />
            </Field>
          </div>
        </div>
      </div>
    </Card>
  );
}
