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

import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ConnectorOrganizationSelect_connector$key } from "#/__generated__/core/ConnectorOrganizationSelect_connector.graphql";
import type { ConnectorOrganizationSelectMutation } from "#/__generated__/core/ConnectorOrganizationSelectMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { connectorCard } from "../variants";

const fragment = graphql`
  fragment ConnectorOrganizationSelect_connector on Connector {
    id
    selectedOrganization
    providerOrganizations {
      status
      nodes {
        slug
        displayName
      }
    }
  }
`;

const configureMutation = graphql`
  mutation ConnectorOrganizationSelectMutation($input: ConfigureConnectorOrganizationInput!) {
    configureConnectorOrganization(input: $input) {
      connector {
        id
        selectedOrganization
        ...ConnectorListItem_connector
      }
    }
  }
`;

interface ConnectorOrganizationSelectProps {
  connectorKey: ConnectorOrganizationSelect_connector$key;
}

export function ConnectorOrganizationSelect({
  connectorKey,
}: ConnectorOrganizationSelectProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const connector = useFragment(fragment, connectorKey);
  const [configure, isUpdating] = useMutation<ConnectorOrganizationSelectMutation>(
    configureMutation,
    {
      successMessage: t("listPage.messages.organizationUpdated"),
      errorToast: t("listPage.errors.organization"),
    },
  );
  const { organizationSelect } = connectorCard();
  const organizations = connector.providerOrganizations;

  if (organizations.status === "NOT_APPLICABLE") {
    return null;
  }

  const options = [...organizations.nodes];
  if (
    connector.selectedOrganization
    && !options.some(option => option.slug === connector.selectedOrganization)
  ) {
    options.unshift({
      slug: connector.selectedOrganization,
      displayName: connector.selectedOrganization,
    });
  }

  function handleChange(value: string | null) {
    if (value == null || value === connector.selectedOrganization) {
      return;
    }

    void configure({
      variables: {
        input: {
          connectorId: connector.id,
          organizationSlug: value,
        },
      },
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <div className={organizationSelect()}>
      <Select
        value={connector.selectedOrganization}
        disabled={isUpdating || organizations.status !== "AVAILABLE"}
        onValueChange={handleChange}
      >
        <SelectTrigger size={1} aria-label={t("listPage.organization")}>
          {(value: string | null) => {
            if (value == null) {
              return (
                <Text size={1} color="faint">
                  {t("listPage.organizationPlaceholder")}
                </Text>
              );
            }

            return options.find(option => option.slug === value)?.displayName ?? value;
          }}
        </SelectTrigger>
        <SelectPopup align="end">
          {options.map(option => (
            <SelectItem key={option.slug} value={option.slug}>
              {option.displayName}
            </SelectItem>
          ))}
        </SelectPopup>
      </Select>
    </div>
  );
}
