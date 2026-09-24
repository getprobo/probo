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

import { formatDatetime, toDateInput } from "@probo/helpers";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { useUpdateUserMutation } from "#/__generated__/iam/useUpdateUserMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const updateUserMutation = graphql`
  mutation useUpdateUserMutation($input: UpdateUserInput!) {
    updateUser(input: $input) {
      profile {
        id
        fullName
        kind
        position
        additionalEmailAddresses
        contract {
          start
          end
        }
      }
    }
  }
`;

export interface UserUpdateSnapshot {
  id: string;
  fullName: string;
  kind: string | null | undefined;
  position: string | null | undefined;
  additionalEmailAddresses: readonly string[];
  contract: {
    start: string | null | undefined;
    end: string | null | undefined;
  } | null | undefined;
}

export function emptyToNull(value: string | null | undefined) {
  if (value == null || value.trim() === "") {
    return null;
  }
  return value;
}

export function useUpdateUser() {
  const { t } = useTranslation();
  return useMutation<useUpdateUserMutation>(updateUserMutation, {
    successMessage: t("userForm.messages.updated"),
    errorToast: t("userForm.errors.update"),
  });
}

export function userUpdateInput(
  profile: UserUpdateSnapshot,
  patch: {
    fullName?: string;
    kind?: string | null;
    position?: string | null;
    additionalEmailAddresses?: readonly string[];
    contractStart?: string | null;
    contractEnd?: string | null;
  },
) {
  const fullName = patch.fullName ?? profile.fullName;
  const kind = patch.kind !== undefined ? patch.kind : profile.kind;
  const position = patch.position !== undefined ? patch.position : profile.position;
  const additionalEmailAddresses = patch.additionalEmailAddresses
    ?? profile.additionalEmailAddresses;
  const contractStart = patch.contractStart !== undefined
    ? patch.contractStart
    : toDateInput(profile.contract?.start);
  const contractEnd = patch.contractEnd !== undefined
    ? patch.contractEnd
    : toDateInput(profile.contract?.end);

  return {
    id: profile.id,
    fullName,
    kind: emptyToNull(kind),
    position: emptyToNull(position),
    additionalEmailAddresses: additionalEmailAddresses.filter(email => email.trim() !== ""),
    contract: {
      start: formatDatetime(emptyToNull(contractStart)) ?? null,
      end: formatDatetime(emptyToNull(contractEnd)) ?? null,
    },
  };
}
