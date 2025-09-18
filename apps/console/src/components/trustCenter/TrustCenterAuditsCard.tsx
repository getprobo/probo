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
  IconCheckmark1,
  IconCrossLargeX,
  Badge,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useFragment } from "react-relay";
import { useMemo, useState } from "react";
import { sprintf, getAuditStateVariant, getAuditStateLabel, formatDate } from "@probo/helpers";
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
    showOnTrustCenter
    createdAt
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      id: string;
      showOnTrustCenter: boolean;
    } & Params;
  };
}) => void;

type Props<Params> = {
  audits: TrustCenterAuditsCardFragment$key[];
  params: Params;
  disabled?: boolean;
  onToggleVisibility: Mutation<Params>;
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

  const onToggleVisibility = (auditId: string, showOnTrustCenter: boolean) => {
    props.onToggleVisibility({
      variables: {
        input: {
          id: auditId,
          showOnTrustCenter,
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
            <Th></Th>
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
              onToggleVisibility={onToggleVisibility}
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
  onToggleVisibility: (auditId: string, showOnTrustCenter: boolean) => void;
  disabled?: boolean;
}) {
  const audit = useFragment(trustCenterAuditFragment, props.audit);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

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
      <Td>
        <div className="flex items-center gap-2">
          {audit.showOnTrustCenter ? (
            <>
              <IconCheckmark1 className="w-4 h-4 text-txt-primary" />
              <span className="text-txt-primary">{__("Visible")}</span>
            </>
          ) : (
            <>
              <IconCrossLargeX className="w-4 h-4 text-txt-tertiary" />
              <span className="text-txt-tertiary">{__("Hidden")}</span>
            </>
          )}
        </div>
      </Td>
      <Td noLink width={100} className="text-end">
        <Button
          variant="secondary"
          onClick={() => props.onToggleVisibility(audit.id, !audit.showOnTrustCenter)}
          icon={audit.showOnTrustCenter ? IconCrossLargeX : IconCheckmark1}
          disabled={props.disabled}
        >
          {audit.showOnTrustCenter ? __("Hide") : __("Show")}
        </Button>
      </Td>
    </Tr>
  );
}
