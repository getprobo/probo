// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import { Field } from "@probo/ui";
import { useCallback, useEffect, useState } from "react";
import { graphql } from "relay-runtime";

import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const setTrustCenterAliasMutation = graphql`
  mutation CompliancePageAliasField_setTrustCenterAliasMutation(
    $input: SetTrustCenterAliasInput!
  ) {
    setTrustCenterAlias(input: $input) {
      alias {
        resourceId
        alias
      }
    }
  }
`;

const removeTrustCenterAliasMutation = graphql`
  mutation CompliancePageAliasField_removeTrustCenterAliasMutation(
    $input: RemoveTrustCenterAliasInput!
  ) {
    removeTrustCenterAlias(input: $input) {
      deletedResourceId
    }
  }
`;

export function CompliancePageAliasField(props: {
  resourceId: string;
  alias: string | null | undefined;
  canSetAlias: boolean;
  canRemoveAlias: boolean;
}) {
  const { resourceId, alias, canSetAlias, canRemoveAlias } = props;

  const { __ } = useTranslate();
  const [value, setValue] = useState(alias ?? "");

  useEffect(() => {
    setValue(alias ?? "");
  }, [alias]);

  const [setTrustCenterAlias, isSettingAlias] = useMutationWithToasts(
    setTrustCenterAliasMutation,
    {
      successMessage: __("Alias updated successfully."),
      errorMessage: __("Failed to update alias"),
    },
  );
  const [removeTrustCenterAlias, isRemovingAlias] = useMutationWithToasts(
    removeTrustCenterAliasMutation,
    {
      successMessage: __("Alias removed successfully."),
      errorMessage: __("Failed to remove alias"),
    },
  );

  const handleBlur = useCallback(async () => {
    const trimmed = value.trim();
    const current = alias ?? "";

    if (trimmed === current) {
      return;
    }

    try {
      if (trimmed === "") {
        if (current !== "" && canRemoveAlias) {
          await removeTrustCenterAlias({
            variables: {
              input: {
                resourceId,
              },
            },
          });
        }

        return;
      }

      if (!canSetAlias) {
        setValue(current);
        return;
      }

      await setTrustCenterAlias({
        variables: {
          input: {
            resourceId,
            alias: trimmed,
          },
        },
      });
    } catch {
      // useMutationWithToasts already shows an error toast.
    }
  }, [alias, canRemoveAlias, canSetAlias, removeTrustCenterAlias, resourceId, setTrustCenterAlias, value]);

  const canEdit = canSetAlias || (canRemoveAlias && (alias ?? "") !== "");

  return (
    <Field
      type="text"
      value={value}
      onChange={event => setValue(event.target.value)}
      onBlur={() => void handleBlur()}
      disabled={!canEdit || isSettingAlias || isRemovingAlias}
      placeholder={__("e.g. privacy-policy")}
      className="min-w-[160px]"
    />
  );
}
