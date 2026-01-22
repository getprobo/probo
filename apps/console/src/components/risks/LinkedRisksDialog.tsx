import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconMagnifyingGlass,
  IconPlusLarge,
  IconTrashCan,
  Input,
  Option,
  Select,
  Spinner,
} from "@probo/ui";
import { type ReactNode, Suspense, useMemo, useState } from "react";
import { useLazyLoadQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { LinkedRisksDialogQuery } from "#/__generated__/core/LinkedRisksDialogQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const risksQuery = graphql`
  query LinkedRisksDialogQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        risks(first: 100) {
          edges {
            node {
              id
              name
              category
              description
              inherentRiskScore
              residualRiskScore
            }
          }
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  connectionId: string;
  disabled?: boolean;
  linkedRisks?: { id: string }[];
  onLink: (riskId: string) => void;
  onUnlink: (riskId: string) => void;
};

export function LinkedRisksDialog({ children, ...props }: Props) {
  const { __ } = useTranslate();

  return (
    <Dialog trigger={children} title={__("Link risks")}>
      <DialogContent>
        <Suspense fallback={<Spinner centered />}>
          <LinkedRisksDialogContent {...props} />
        </Suspense>
      </DialogContent>
      <DialogFooter exitLabel={__("Close")} />
    </Dialog>
  );
}

function LinkedRisksDialogContent(props: Omit<Props, "children">) {
  const organizationId = useOrganizationId();
  const data = useLazyLoadQuery<LinkedRisksDialogQuery>(risksQuery, {
    organizationId,
  });
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState<string | null>(null);
  const risks = useMemo(
    () => data.organization?.risks?.edges?.map(edge => edge.node) ?? [],
    [data.organization?.risks],
  );
  const linkedIds = useMemo(() => {
    return new Set(props.linkedRisks?.map(r => r.id) ?? []);
  }, [props.linkedRisks]);

  const filteredRisks = useMemo(() => {
    return risks.filter(
      risk =>
        (category === null || risk.category === category)
        && (risk.name.toLowerCase().includes(search.toLowerCase())
          || risk.description?.toLowerCase().includes(search.toLowerCase())),
    );
  }, [risks, search, category]);

  const categories = useMemo(
    () => Array.from(new Set(risks.map(r => r.category))),
    [risks],
  );

  return (
    <>
      <div className="flex items-center gap-2 sticky top-0 relative py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0 px-6">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={__("Search risks...")}
          onValueChange={setSearch}
        />
        <Select
          value={category ?? ""}
          placeholder={__("All categories")}
          onValueChange={setCategory}
          className="max-w-[180px]"
        >
          {categories.map(category => (
            <Option key={category} value={category}>
              {category}
            </Option>
          ))}
        </Select>
      </div>
      <div className="divide-y divide-border-low">
        {filteredRisks.map(risk => (
          <RiskRow
            key={risk.id}
            risk={risk}
            linkedRisks={linkedIds}
            onLink={props.onLink}
            onUnlink={props.onUnlink}
            disabled={props.disabled}
          />
        ))}
      </div>
    </>
  );
}

type RowProps = {
  risk: {
    name: string;
    category: string;
    id: string;
    inherentRiskScore: number;
    residualRiskScore: number;
  };
  linkedRisks: Set<string>;
  disabled?: boolean;
  onLink: (riskId: string) => void;
  onUnlink: (riskId: string) => void;
};

function RiskRow(props: RowProps) {
  const { __ } = useTranslate();

  const isLinked = props.linkedRisks.has(props.risk.id);
  const onClick = isLinked ? props.onUnlink : props.onLink;
  const IconComponent = isLinked ? IconTrashCan : IconPlusLarge;

  return (
    <button
      className="py-4 flex items-center gap-4 hover:bg-subtle cursor-pointer px-6 w-full"
      onClick={() => onClick(props.risk.id)}
    >
      <div className="text-left">{props.risk.name}</div>
      <Badge variant="neutral">{props.risk.category}</Badge>
      <Button
        disabled={props.disabled}
        className="ml-auto"
        variant={isLinked ? "secondary" : "primary"}
        asChild
      >
        <span>
          <IconComponent size={16} />
          {" "}
          {isLinked ? __("Unlink") : __("Link")}
        </span>
      </Button>
    </button>
  );
}
