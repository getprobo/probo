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
import { FrameworkLogo, IconCheckmark1, Spinner } from "@probo/ui";
import { useCallback, useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageFrameworkListItem_complianceFramework$key } from "#/__generated__/core/CompliancePageFrameworkListItem_complianceFramework.graphql";
import type { CompliancePageFrameworkListItem_compliancePage$key } from "#/__generated__/core/CompliancePageFrameworkListItem_compliancePage.graphql";
import type { CompliancePageFrameworkListItem_createMutation } from "#/__generated__/core/CompliancePageFrameworkListItem_createMutation.graphql";
import type { CompliancePageFrameworkListItem_deleteMutation } from "#/__generated__/core/CompliancePageFrameworkListItem_deleteMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

export const compliancePageFrameworkListItemFragment = graphql`
  fragment CompliancePageFrameworkListItem_complianceFramework on ComplianceFramework {
    id
    visibility
    framework {
      id
      name
      lightLogo {
        downloadUrl
      }
      darkLogo {
        downloadUrl
      }
    }
  }
`;

const compliancePageFragment = graphql`
  fragment CompliancePageFrameworkListItem_compliancePage on CompliancePortal {
    id
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

const createMutation = graphql`
  mutation CompliancePageFrameworkListItem_createMutation($input: CreateComplianceFrameworkInput!) {
    createComplianceFramework(input: $input) {
      complianceFrameworkEdge {
        node {
          id
        }
      }
    }
  }
`;

const deleteMutation = graphql`
  mutation CompliancePageFrameworkListItem_deleteMutation($input: DeleteComplianceFrameworkInput!) {
    deleteComplianceFramework(input: $input) {
      deletedComplianceFrameworkId
    }
  }
`;

export interface CompliancePageFrameworkListItemProps {
  complianceFrameworkKey: CompliancePageFrameworkListItem_complianceFramework$key;
  compliancePageKey: CompliancePageFrameworkListItem_compliancePage$key;
  onRefetch: () => void;
}

export function CompliancePageFrameworkListItem(props: CompliancePageFrameworkListItemProps) {
  const { complianceFrameworkKey, compliancePageKey, onRefetch } = props;

  const { __ } = useTranslate();
  const [optimisticPublic, setOptimisticPublic] = useState<boolean | null>(null);

  const complianceFramework = useFragment(
    compliancePageFrameworkListItemFragment,
    complianceFrameworkKey,
  );
  const compliancePage = useFragment(compliancePageFragment, compliancePageKey);
  const canUpdate = compliancePage.canUpdate;
  const compliancePortalId = compliancePage.id;
  const { id, visibility, framework } = complianceFramework;

  const serverPublic = visibility === "PUBLIC";

  if (optimisticPublic !== null && optimisticPublic === serverPublic) {
    setOptimisticPublic(null);
  }

  const isPublic = optimisticPublic ?? serverPublic;

  const [createComplianceFramework, isCreating] = useMutation<CompliancePageFrameworkListItem_createMutation>(
    createMutation,
    {
      successMessage: __("Framework visibility updated successfully."),
      errorToast: __("Failed to update framework visibility"),
    },
  );

  const [deleteComplianceFramework, isDeleting] = useMutation<CompliancePageFrameworkListItem_deleteMutation>(
    deleteMutation,
    {
      successMessage: __("Framework visibility updated successfully."),
      errorToast: __("Failed to update framework visibility"),
    },
  );

  const isLoading = isCreating || isDeleting;

  const handleToggle = useCallback(async () => {
    if (!canUpdate || isLoading) return;

    const nextPublic = !isPublic;
    setOptimisticPublic(nextPublic);

    try {
      if (isPublic) {
        await deleteComplianceFramework({
          variables: { input: { id } },
          onCompleted: (_, errors) => {
            if (errors?.length) {
              setOptimisticPublic(null);
              return;
            }
            onRefetch();
          },
        });
      } else {
        await createComplianceFramework({
          variables: {
            input: {
              compliancePortalId,
              frameworkId: framework.id,
            },
          },
          onCompleted: (_, errors) => {
            if (errors?.length) {
              setOptimisticPublic(null);
              return;
            }
            onRefetch();
          },
        });
      }
    } catch {
      setOptimisticPublic(null);
    }
  }, [
    canUpdate,
    isLoading,
    isPublic,
    deleteComplianceFramework,
    id,
    onRefetch,
    createComplianceFramework,
    compliancePortalId,
    framework.id,
  ]);

  const className = [
    "relative flex flex-col items-center gap-3 rounded-lg border p-4 text-center",
    "transition-[background-color,border-color,box-shadow,opacity] duration-200 ease-in-out",
    isPublic
      ? "border-primary-500 bg-primary-50 ring-2 ring-primary-500"
      : "border-border-solid bg-surface-secondary hover:border-border-medium hover:bg-surface-primary",
    canUpdate && !isLoading && "cursor-pointer",
    !canUpdate && "cursor-default",
    isLoading && "pointer-events-none",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <button
      type="button"
      disabled={!canUpdate}
      aria-pressed={isPublic}
      aria-busy={isLoading}
      aria-label={framework.name}
      onClick={() => void handleToggle()}
      className={className}
    >
      <span
        className={[
          "absolute top-2 right-2 flex size-5 items-center justify-center rounded-full bg-primary-500 text-white",
          "transition-[opacity,transform] duration-200 ease-in-out",
          isPublic ? "scale-100 opacity-100" : "scale-75 opacity-0",
        ].join(" ")}
        aria-hidden={!isPublic}
      >
        <IconCheckmark1 size={12} />
      </span>

      <div
        className={[
          "flex flex-col items-center gap-3 transition-opacity duration-200 ease-in-out",
          isLoading ? "opacity-50" : "opacity-100",
        ].join(" ")}
      >
        <FrameworkLogo
          className="size-12"
          lightLogoURL={framework.lightLogo?.downloadUrl}
          darkLogoURL={framework.darkLogo?.downloadUrl}
          name={framework.name}
        />

        <span className="text-sm font-medium">{framework.name}</span>
      </div>

      {isLoading && (
        <span className="absolute inset-0 flex items-center justify-center rounded-lg">
          <Spinner size={24} />
        </span>
      )}
    </button>
  );
}
