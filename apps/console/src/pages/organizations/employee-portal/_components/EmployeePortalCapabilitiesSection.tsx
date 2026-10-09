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

import { LaptopIcon } from "@phosphor-icons/react";
import { Switch } from "@probo/ui/src/v2/Switch/Switch";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { EmployeePortalCapabilitiesSection_employeePortal$key } from "#/__generated__/core/EmployeePortalCapabilitiesSection_employeePortal.graphql";
import type { EmployeePortalCapabilitiesSection_updateMutation } from "#/__generated__/core/EmployeePortalCapabilitiesSection_updateMutation.graphql";
import { TonedCard } from "#/components/TonedCard/TonedCard";
import { useMutation } from "#/lib/relay/useMutation";

import { employeePortalCapabilitiesSection } from "../variants";

const NS = "organizations/employee-portal";

const employeePortalFragment = graphql`
  fragment EmployeePortalCapabilitiesSection_employeePortal on EmployeePortal {
    id
    capabilities {
      deviceAgent
    }
    canUpdate: permission(action: "employee-portal:portal:update")
  }
`;

const updateMutation = graphql`
  mutation EmployeePortalCapabilitiesSection_updateMutation(
    $input: UpdateEmployeePortalInput!
  ) {
    updateEmployeePortal(input: $input) {
      employeePortal {
        id
        capabilities {
          deviceAgent
        }
      }
    }
  }
`;

export interface EmployeePortalCapabilitiesSectionProps {
  employeePortalKey: EmployeePortalCapabilitiesSection_employeePortal$key;
}

export function EmployeePortalCapabilitiesSection({
  employeePortalKey,
}: EmployeePortalCapabilitiesSectionProps) {
  const { t } = useTranslation(NS);
  const { root, intro } = employeePortalCapabilitiesSection();
  const employeePortal = useFragment(employeePortalFragment, employeePortalKey);
  const checked = employeePortal.capabilities.deviceAgent;

  const [updatePortal, isUpdating]
    = useMutation<EmployeePortalCapabilitiesSection_updateMutation>(
      updateMutation,
      {
        successMessage: t("capabilities.messages.updated"),
        errorToast: t("capabilities.errors.update"),
      },
    );

  const disabled = isUpdating || !employeePortal.canUpdate;
  const title = t("capabilities.deviceAgent.title");

  function handleCheckedChange(next: boolean) {
    void updatePortal({
      variables: {
        input: {
          employeePortalId: employeePortal.id,
          capabilities: {
            deviceAgent: next,
          },
        },
      },
    });
  }

  return (
    <section className={root()}>
      <div className={intro()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("capabilities.title")}
        </Heading>
      </div>
      <TonedCard
        tone={checked ? "green" : "sand"}
        icon={<LaptopIcon size={20} weight="duotone" />}
        control={(
          <Switch
            size={2}
            checked={checked}
            color="green"
            disabled={disabled}
            aria-label={title}
            onCheckedChange={handleCheckedChange}
          />
        )}
      >
        <Text size={3} weight="medium" highContrast>
          {title}
        </Text>
        <Text size={2} color="neutral">
          {t("capabilities.deviceAgent.description")}
        </Text>
      </TonedCard>
    </section>
  );
}
