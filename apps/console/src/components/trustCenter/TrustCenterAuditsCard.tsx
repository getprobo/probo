import { graphql } from "relay-runtime";
import {
  Card,
  Button,
  Tr,
  Td,
  Table,
  Thead,
  Tbody,
  Th,
  IconChevronDown,
  Badge,
  Field,
  Option,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useFragment } from "react-relay";
import { useMemo, useState, useCallback, useEffect } from "react";
import { sprintf, getAuditStateVariant, getAuditStateLabel, formatDate, getTrustCenterVisibilityOptions } from "@probo/helpers";
import { useOrganizationId } from "/hooks/useOrganizationId";
import clsx from "clsx";
import type { TrustCenterAuditsCardFragment$key } from "./__generated__/TrustCenterAuditsCardFragment.graphql";

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
}) => void;

type Props<Params> = {
  audits: TrustCenterAuditsCardFragment$key[];
  params: Params;
  disabled?: boolean;
  onChangeVisibility: Mutation<Params>;
  variant?: "card" | "table";
};

export function TrustCenterAuditsCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const [limit, setLimit] = useState<number | null>(4);
  const audits = useMemo(() => {
    return limit ? props.audits.slice(0, limit) : props.audits;
  }, [props.audits, limit]);
  const showMoreButton = limit !== null && props.audits.length > limit;
  const variant = props.variant ?? "table";

  const onChangeVisibility = (auditId: string, trustCenterVisibility: "NONE" | "PRIVATE" | "PUBLIC") => {
    props.onChangeVisibility({
      variables: {
        input: {
          id: auditId,
          trustCenterVisibility,
          ...props.params,
        },
      },
    });
  };

  const Wrapper = variant === "card" ? Card : "div";

  return (
    <Wrapper {...(variant === "card" ? { padded: true } : {})} className="space-y-[10px]">
      <Table className={clsx(variant === "card" && "bg-invert")}>
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
              audit={audit}
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
    </Wrapper>
  );
}

function AuditRow(props: {
  audit: TrustCenterAuditsCardFragment$key;
  onChangeVisibility: (auditId: string, trustCenterVisibility: "NONE" | "PRIVATE" | "PUBLIC") => void;
  disabled?: boolean;
}) {
  const audit = useFragment(trustCenterAuditFragment, props.audit);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const [optimisticValue, setOptimisticValue] = useState<string | null>(null);

  const handleValueChange = useCallback((value: string | {}) => {
    const stringValue = typeof value === 'string' ? value : '';
    const typedValue = stringValue as "NONE" | "PRIVATE" | "PUBLIC";
    setOptimisticValue(typedValue);
    props.onChangeVisibility(audit.id, typedValue);
  }, [audit.id, props.onChangeVisibility]);

  useEffect(() => {
    if (optimisticValue && audit.trustCenterVisibility === optimisticValue) {
      setOptimisticValue(null);
    }
  }, [audit.trustCenterVisibility, optimisticValue]);

  const currentValue = optimisticValue || audit.trustCenterVisibility;

  const visibilityOptions = getTrustCenterVisibilityOptions(__);

  const validUntilFormatted = audit.validUntil
    ? formatDate(audit.validUntil)
    : __("No expiry");

  return (
    <Tr to={`/organizations/${organizationId}/audits/${audit.id}`}>
      <Td>
        <div className="flex gap-4 items-center">
          {audit.framework.name}
        </div>
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
          value={currentValue}
          onValueChange={handleValueChange}
          disabled={props.disabled}
          className="w-[105px]"
        >
          {visibilityOptions.map((option) => (
            <Option key={option.value} value={option.value}>
              <div className="flex items-center justify-between w-full">
                <Badge variant={option.variant}>
                  {option.label}
                </Badge>
              </div>
            </Option>
          ))}
        </Field>
      </Td>
    </Tr>
  );
}
