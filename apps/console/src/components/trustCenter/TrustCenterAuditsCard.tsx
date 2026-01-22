import {
  formatDate,
  getAuditStateLabel,
  getAuditStateVariant,
  getTrustCenterVisibilityOptions,
  sprintf,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Field,
  IconChevronDown,
  Option,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useCallback, useMemo, useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { TrustCenterAuditsCardFragment$key } from "#/__generated__/core/TrustCenterAuditsCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const trustCenterAuditFragment = graphql`
  fragment TrustCenterAuditsCardFragment on Audit {
    id
    name
    framework {
      name
    }
    validFrom
    validUntil
    state
    trustCenterVisibility
    createdAt
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      id: string;
      trustCenterVisibility: "NONE" | "PRIVATE" | "PUBLIC";
    } & Params;
  };
}) => Promise<void>;

type Props<Params> = {
  audits: TrustCenterAuditsCardFragment$key[];
  params: Params;
  disabled?: boolean;
  onChangeVisibility: Mutation<Params>;
  canUpdate: boolean;
};

export function TrustCenterAuditsCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const [limit, setLimit] = useState<number | null>(100);
  const audits = useMemo(() => {
    return limit ? props.audits.slice(0, limit) : props.audits;
  }, [props.audits, limit]);
  const showMoreButton = limit !== null && props.audits.length > limit;

  const onChangeVisibility = async (
    auditId: string,
    trustCenterVisibility: "NONE" | "PRIVATE" | "PUBLIC",
  ) => {
    await props.onChangeVisibility({
      variables: {
        input: {
          id: auditId,
          trustCenterVisibility,
          ...props.params,
        },
      },
    });
  };

  return (
    <div className="space-y-[10px]">
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Framework")}</Th>
            <Th>{__("Name")}</Th>
            <Th>{__("Valid Until")}</Th>
            <Th>{__("State")}</Th>
            <Th>{__("Visibility")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {audits.length === 0 && (
            <Tr>
              <Td colSpan={6} className="text-center text-txt-secondary">
                {__("No audits available")}
              </Td>
            </Tr>
          )}
          {audits.map((audit, index) => (
            <AuditRow
              key={index}
              auditFragmentRef={audit}
              canUpdate={props.canUpdate}
              onChangeVisibility={onChangeVisibility}
              disabled={props.disabled}
            />
          ))}
        </Tbody>
      </Table>
      {showMoreButton && (
        <Button
          variant="tertiary"
          onClick={() => setLimit(null)}
          className="mt-3 mx-auto"
          icon={IconChevronDown}
        >
          {sprintf(__("Show %s more"), props.audits.length - limit)}
        </Button>
      )}
    </div>
  );
}

function AuditRow(props: {
  auditFragmentRef: TrustCenterAuditsCardFragment$key;
  onChangeVisibility: (
    auditId: string,
    trustCenterVisibility: "NONE" | "PRIVATE" | "PUBLIC",
  ) => Promise<void>;
  disabled?: boolean;
  canUpdate: boolean;
}) {
  const { auditFragmentRef, onChangeVisibility, disabled, canUpdate } = props;
  const audit = useFragment(trustCenterAuditFragment, auditFragmentRef);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const handleValueChange = useCallback(
    async (value: string) => {
      const stringValue = typeof value === "string" ? value : "";
      const typedValue = stringValue as "NONE" | "PRIVATE" | "PUBLIC";
      await onChangeVisibility(audit.id, typedValue);
    },
    [audit.id, onChangeVisibility],
  );

  const visibilityOptions = getTrustCenterVisibilityOptions(__);

  const validUntilFormatted = audit.validUntil
    ? formatDate(audit.validUntil)
    : __("No expiry");

  return (
    <Tr to={`/organizations/${organizationId}/audits/${audit.id}`}>
      <Td>
        <div className="flex gap-4 items-center">{audit.framework.name}</div>
      </Td>
      <Td>{audit.name || __("Untitled")}</Td>
      <Td>{validUntilFormatted}</Td>
      <Td>
        <Badge variant={getAuditStateVariant(audit.state)}>
          {getAuditStateLabel(__, audit.state)}
        </Badge>
      </Td>
      <Td noLink width={130} className="pr-0">
        <Field
          type="select"
          value={audit.trustCenterVisibility}
          onValueChange={value => void handleValueChange(value)}
          disabled={disabled || !canUpdate}
          className="w-[105px]"
        >
          {visibilityOptions.map(option => (
            <Option key={option.value} value={option.value}>
              <div className="flex items-center justify-between w-full">
                <Badge variant={option.variant}>{option.label}</Badge>
              </div>
            </Option>
          ))}
        </Field>
      </Td>
    </Tr>
  );
}
