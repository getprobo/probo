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

import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconMagnifyingGlass,
  IconPlusLarge,
  IconTrashCan,
  InfiniteScrollTrigger,
  Input,
  Spinner,
} from "@probo/ui";
import { cloneElement, isValidElement, type MouseEvent, type ReactElement, type ReactNode, Suspense, useMemo, useState } from "react";
import {
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { graphql } from "relay-runtime";

import type {
  LinkScenarioDialogFragment$data,
  LinkScenarioDialogFragment$key,
} from "#/__generated__/core/LinkScenarioDialogFragment.graphql";
import type { LinkScenarioDialogQuery } from "#/__generated__/core/LinkScenarioDialogQuery.graphql";
import type { LinkScenarioDialogQuery_fragment } from "#/__generated__/core/LinkScenarioDialogQuery_fragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

const scenariosQuery = graphql`
  query LinkScenarioDialogQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...LinkScenarioDialogFragment
      }
    }
  }
`;

const scenariosFragment = graphql`
  fragment LinkScenarioDialogFragment on Organization
  @refetchable(queryName: "LinkScenarioDialogQuery_fragment")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    riskAssessmentScenarios(first: $first, after: $after, last: $last, before: $before)
      @connection(key: "LinkScenarioDialogQuery_riskAssessmentScenarios") {
      edges {
        node {
          id
          name
          description
        }
      }
    }
  }
`;

interface LinkScenarioDialogProps {
  children: ReactNode;
  connectionId: string;
  disabled?: boolean;
  linkedScenarios?: { id: string }[];
  onLink: (scenarioId: string) => void;
  onUnlink: (scenarioId: string) => void;
}

export function LinkScenarioDialog({ children, ...props }: LinkScenarioDialogProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery]
    = useQueryLoader<LinkScenarioDialogQuery>(scenariosQuery);

  const trigger = isValidElement(children)
    ? cloneElement(children as ReactElement<{ onClick?: (e: MouseEvent) => void }>, {
        onClick: (e: MouseEvent) => {
          (children as ReactElement<{ onClick?: (e: MouseEvent) => void }>).props.onClick?.(e);
          loadQuery({ organizationId }, { fetchPolicy: "network-only" });
        },
      })
    : children;

  return (
    <Dialog trigger={trigger} title={__("Link scenarios")}>
      <DialogContent>
        {queryRef
          ? (
              <Suspense fallback={<Spinner centered />}>
                <LinkScenarioDialogContent queryRef={queryRef} {...props} />
              </Suspense>
            )
          : (
              <Spinner centered />
            )}
      </DialogContent>
      <DialogFooter exitLabel={__("Close")} />
    </Dialog>
  );
}

type ContentProps = Omit<LinkScenarioDialogProps, "children"> & {
  queryRef: PreloadedQuery<LinkScenarioDialogQuery>;
};

function LinkScenarioDialogContent(props: ContentProps) {
  const query = usePreloadedQuery<LinkScenarioDialogQuery>(scenariosQuery, props.queryRef);
  const { data, loadNext, hasNext, isLoadingNext }
    = usePaginationFragment<LinkScenarioDialogQuery_fragment, LinkScenarioDialogFragment$key>(
      scenariosFragment,
      query.organization as LinkScenarioDialogFragment$key,
    );
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const scenarios = useMemo(
    () => data.riskAssessmentScenarios?.edges?.map(edge => edge.node) ?? [],
    [data.riskAssessmentScenarios],
  );
  const linkedIds = useMemo(() => {
    return new Set(props.linkedScenarios?.map(s => s.id) ?? []);
  }, [props.linkedScenarios]);

  const filteredScenarios = useMemo(() => {
    return scenarios.filter(
      scenario =>
        scenario.name?.toLowerCase().includes(search.toLowerCase())
        || scenario.description?.toLowerCase().includes(search.toLowerCase()),
    );
  }, [scenarios, search]);

  return (
    <>
      <div className="flex items-center gap-2 sticky top-0 relative py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0 px-6">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={__("Search scenarios...")}
          onValueChange={setSearch}
        />
      </div>
      <div className="divide-y divide-border-low">
        {filteredScenarios.map(scenario => (
          <ScenarioRow
            key={scenario.id}
            scenario={scenario}
            linkedScenarios={linkedIds}
            onLink={props.onLink}
            onUnlink={props.onUnlink}
            disabled={props.disabled}
          />
        ))}
        {hasNext && (
          <InfiniteScrollTrigger
            loading={isLoadingNext}
            onView={() => loadNext(20)}
          />
        )}
      </div>
    </>
  );
}

type Scenario = NodeOf<LinkScenarioDialogFragment$data["riskAssessmentScenarios"]>;

function ScenarioRow(props: {
  scenario: Scenario;
  linkedScenarios: Set<string>;
  onLink: (scenarioId: string) => void;
  onUnlink: (scenarioId: string) => void;
  disabled?: boolean;
}) {
  const { __ } = useTranslate();
  const isLinked = props.linkedScenarios.has(props.scenario.id);

  const onToggle = () => {
    if (isLinked) {
      props.onUnlink(props.scenario.id);
    } else {
      props.onLink(props.scenario.id);
    }
  };

  return (
    <div className="flex items-center justify-between p-4 hover:bg-level-1">
      <div className="flex-1 min-w-0">
        <div className="text-sm font-medium text-txt-primary truncate">
          {props.scenario.name}
        </div>
        <div className="text-xs text-txt-secondary truncate">
          {props.scenario.description || __("No description")}
        </div>
      </div>
      <Button
        variant={isLinked ? "secondary" : "primary"}
        icon={isLinked ? IconTrashCan : IconPlusLarge}
        onClick={onToggle}
        disabled={props.disabled}
        className="ml-6"
      >
        {isLinked ? __("Unlink") : __("Link")}
      </Button>
    </div>
  );
}
