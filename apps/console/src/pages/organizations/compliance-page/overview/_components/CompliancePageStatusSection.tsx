// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
import { Card, Spinner, Toggle, useToast } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageStatusSectionFragment$key } from "#/__generated__/core/CompliancePageStatusSectionFragment.graphql";
import { useUpdateTrustCenterMutation } from "#/hooks/graph/TrustCenterGraph";

const fragment = graphql`
  fragment CompliancePageStatusSectionFragment on Organization {
    compliancePage: trustCenter {
      id
      active
      searchEngineIndexing
      canUpdate: permission(action: "core:trust-center:update")
    }
  }
`;

export function CompliancePageStatusSection(props: {
  fragmentRef: CompliancePageStatusSectionFragment$key;
}) {
  const { fragmentRef } = props;

  const { __ } = useTranslate();
  const { toast } = useToast();

  const organization = useFragment<CompliancePageStatusSectionFragment$key>(
    fragment,
    fragmentRef,
  );

  const [updateCompliancePage, isUpdating] = useUpdateTrustCenterMutation();

  const handleToggleActive = async (active: boolean) => {
    if (!organization.compliancePage?.id) {
      toast({
        title: __("Error"),
        description: __("Compliance page not found"),
        variant: "error",
      });
      return;
    }

    await updateCompliancePage({
      variables: {
        input: {
          trustCenterId: organization.compliancePage.id,
          active,
        },
      },
    });
  };

  const handleToggleSearchEngineIndexing = async (indexable: boolean) => {
    if (!organization.compliancePage?.id) {
      toast({
        title: __("Error"),
        description: __("Compliance page not found"),
        variant: "error",
      });
      return;
    }

    await updateCompliancePage({
      variables: {
        input: {
          trustCenterId: organization.compliancePage.id,
          searchEngineIndexing: indexable ? "INDEXABLE" : "NOT_INDEXABLE",
        },
      },
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-medium">
          {__("Compliance Page Status")}
        </h2>
        {isUpdating && <Spinner />}
      </div>
      <Card padded className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h3 className="font-medium">{__("Activate Compliance Page")}</h3>
            <p className="text-sm text-txt-tertiary">
              {__(
                "Make your compliance page publicly accessible to build customer confidence",
              )}
            </p>
          </div>
          <Toggle
            checked={!!organization.compliancePage?.active}
            onChange={checked => void handleToggleActive(checked)}
            disabled={!organization.compliancePage?.canUpdate}
          />
        </div>

        <div className="flex items-center justify-between border-t border-border-solid pt-4">
          <div className="space-y-1">
            <h3 className="font-medium">{__("Search Engine Indexing")}</h3>
            <p className="text-sm text-txt-tertiary">
              {__(
                "Allow search engines to index your compliance page and make it discoverable",
              )}
            </p>
          </div>
          <span
            title={
              !organization.compliancePage?.active
                ? __("Activate your compliance page first to enable search engine indexing")
                : undefined
            }
          >
            <Toggle
              checked={
                organization.compliancePage?.searchEngineIndexing === "INDEXABLE"
              }
              onChange={checked =>
                void handleToggleSearchEngineIndexing(checked)}
              disabled={
                !organization.compliancePage?.canUpdate
                || !organization.compliancePage?.active
              }
            />
          </span>
        </div>
      </Card>
    </div>
  );
}
