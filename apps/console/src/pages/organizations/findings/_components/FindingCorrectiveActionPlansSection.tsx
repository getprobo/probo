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

import { PlusIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment } from "react-relay";

import type { FindingCorrectiveActionPlansSection_finding$key } from "#/__generated__/core/FindingCorrectiveActionPlansSection_finding.graphql";
import type { FindingCorrectiveActionPlansSectionCreateMutation } from "#/__generated__/core/FindingCorrectiveActionPlansSectionCreateMutation.graphql";
import type { FindingCorrectiveActionPlansSectionPaginationQuery } from "#/__generated__/core/FindingCorrectiveActionPlansSectionPaginationQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { FindingCorrectiveActionPlanItem } from "./FindingCorrectiveActionPlanItem";
import { correctiveActionPlans, findingsPage } from "./variants";

const findingCorrectiveActionPlansSectionFragment = graphql`
  fragment FindingCorrectiveActionPlansSection_finding on Finding
  @refetchable(queryName: "FindingCorrectiveActionPlansSectionPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey" }
  ) {
    id
    canCreatePlan: permission(action: "core:corrective-action-plan:create")
    correctiveActionPlans(first: $first, after: $after)
      @connection(key: "FindingCorrectiveActionPlansSection_correctiveActionPlans") {
      __id
      edges {
        node {
          id
          ...FindingCorrectiveActionPlanItem_plan
        }
      }
    }
  }
`;

const createPlanMutation = graphql`
  mutation FindingCorrectiveActionPlansSectionCreateMutation(
    $input: CreateCorrectiveActionPlanInput!
    $connections: [ID!]!
  ) {
    createCorrectiveActionPlan(input: $input) {
      correctiveActionPlanEdge @appendEdge(connections: $connections) {
        node {
          id
          ...FindingCorrectiveActionPlanItem_plan
        }
      }
    }
  }
`;

interface FindingCorrectiveActionPlansSectionProps {
  findingKey: FindingCorrectiveActionPlansSection_finding$key;
}

export function FindingCorrectiveActionPlansSection({
  findingKey,
}: FindingCorrectiveActionPlansSectionProps) {
  const { t } = useTranslation("organizations/findings");
  const {
    data: finding,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    FindingCorrectiveActionPlansSectionPaginationQuery,
    FindingCorrectiveActionPlansSection_finding$key
  >(findingCorrectiveActionPlansSectionFragment, findingKey);
  const [createPlan, isCreating] = useMutation<FindingCorrectiveActionPlansSectionCreateMutation>(
    createPlanMutation,
    { errorToast: t("details.plans.createError") },
  );
  const [title, setTitle] = useState("");
  const { section } = findingsPage();
  const { plans, createForm } = correctiveActionPlans();
  const connectionId = finding.correctiveActionPlans.__id;
  const items = finding.correctiveActionPlans.edges.map(edge => edge.node);

  async function submit() {
    const trimmed = title.trim();
    if (trimmed === "") {
      return;
    }
    try {
      await createPlan({
        variables: {
          input: { findingId: finding.id, title: trimmed },
          connections: [connectionId],
        },
      });
      setTitle("");
    } catch {
      // The mutation hook already reported the failure.
    }
  }

  return (
    <Card variant="soft" size={2}>
      <div className={section()}>
        <Heading level={2} size={2} weight="medium">
          {t("details.plans.title")}
        </Heading>
        <Text size={1} color="faint">{t("details.plans.hint")}</Text>
        {items.length === 0
          ? <Text size={2} color="faint">{t("details.plans.empty")}</Text>
          : (
              <div className={plans()}>
                {items.map(plan => (
                  <FindingCorrectiveActionPlanItem
                    key={plan.id}
                    planKey={plan}
                    connectionId={connectionId}
                  />
                ))}
              </div>
            )}
        {hasNext && (
          <Button
            variant="soft"
            color="neutral"
            disabled={isLoadingNext}
            onClick={() => loadNext(20)}
          >
            {t("details.plans.loadMore")}
          </Button>
        )}
        {finding.canCreatePlan && (
          <form
            className={createForm()}
            onSubmit={(event) => {
              event.preventDefault();
              void submit();
            }}
          >
            <TextField
              className="min-w-0 flex-1"
              value={title}
              aria-label={t("details.plans.newTitle")}
              placeholder={t("details.plans.newTitle")}
              disabled={isCreating}
              onValueChange={setTitle}
            />
            <Button
              type="submit"
              variant="soft"
              iconStart={<PlusIcon />}
              disabled={title.trim() === ""}
              loading={isCreating}
            >
              {t("details.plans.add")}
            </Button>
          </form>
        )}
      </div>
    </Card>
  );
}
