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

import { CaretLeftIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { TaskDetails } from "@probo/tasks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Navigate, useLocation, useNavigate } from "react-router";

import type { NewFindingPageCreateMutation } from "#/__generated__/core/NewFindingPageCreateMutation.graphql";
import type { NewFindingPageQuery } from "#/__generated__/core/NewFindingPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";

import { FindingPropertiesDate } from "./_components/FindingPropertiesDate";
import { FindingPropertiesKind } from "./_components/FindingPropertiesKind";
import { FindingPropertiesLayout } from "./_components/FindingPropertiesLayout";
import { type FindingOwnerValue, FindingPropertiesOwner } from "./_components/FindingPropertiesOwner";
import { FindingPropertiesPriority } from "./_components/FindingPropertiesPriority";
import { FindingPropertiesRisk } from "./_components/FindingPropertiesRisk";
import { FindingPropertiesStatus } from "./_components/FindingPropertiesStatus";
import { findingsPage, findingTitleField } from "./_components/variants";
import {
  findingDueIsOverdue,
  type FindingKind,
  type FindingPriority,
  type FindingStatus,
} from "./_lib/findingOptions";
import { findingDetailsPath, findingListPath } from "./_lib/findingPath";
import { clampFindingTitle } from "./_lib/findingTitle";

const createFindingMutation = graphql`
  mutation NewFindingPageCreateMutation($input: CreateFindingInput!) {
    createFinding(input: $input) {
      findingEdge {
        node {
          id
        }
      }
    }
  }
`;

export const newFindingPageQuery = graphql`
  query NewFindingPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canCreateFinding: permission(action: "core:finding:create")
      }
    }
  }
`;

interface NewFindingPageProps {
  queryRef: PreloadedQuery<NewFindingPageQuery>;
}

export function NewFindingPage({ queryRef }: NewFindingPageProps) {
  const data = usePreloadedQuery<NewFindingPageQuery>(newFindingPageQuery, queryRef);
  const { t } = useTranslation("organizations/findings");
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const location = useLocation();
  const [kind, setKind] = useState<FindingKind>("MINOR_NONCONFORMITY");
  const [status, setStatus] = useState<FindingStatus>("OPEN");
  const [priority, setPriority] = useState<FindingPriority>("MEDIUM");
  const [owner, setOwner] = useState<FindingOwnerValue | null>(null);
  const [risk, setRisk] = useState<{ id: string; name: string } | null>(null);
  const [identifiedOn, setIdentifiedOn] = useState<string | null>(null);
  const [dueDate, setDueDate] = useState<string | null>(null);
  const [riskError, setRiskError] = useState<string | null>(null);
  const { back, narrative, narrativeWide } = findingsPage();
  const { input: titleInput } = findingTitleField();
  const [createFinding, isCreating] = useMutation<NewFindingPageCreateMutation>(
    createFindingMutation,
    {
      successMessage: t("list.messages.created"),
      errorToast: t("list.errors.create"),
    },
  );

  usePageTitle(t("create.title"));

  if (data.organization.__typename !== "Organization") {
    throw new NotFoundError("invalid type for organization node");
  }

  if (!data.organization.canCreateFinding) {
    return (
      <Navigate
        to={{ pathname: findingListPath(organizationId), search: location.search }}
        replace
      />
    );
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (status === "RISK_ACCEPTED" && risk == null) {
      setRiskError(t("details.errors.riskRequired"));
      return;
    }
    const formData = new FormData(event.currentTarget);
    const text = (name: string) => {
      const value = formData.get(name);
      if (typeof value !== "string") {
        return null;
      }
      const trimmed = value.trim();
      return trimmed === "" ? null : trimmed;
    };
    const title = formData.get("title");

    try {
      const response = await createFinding({
        variables: {
          input: {
            organizationId,
            kind,
            title: typeof title === "string" ? title.trim() : "",
            description: text("description"),
            source: text("source"),
            rootCause: text("rootCause"),
            correctiveAction: text("correctiveAction"),
            effectivenessCheck: text("effectivenessCheck"),
            status,
            priority,
            ownerId: owner?.id ?? null,
            riskId: risk?.id ?? null,
            identifiedOn,
            dueDate,
          },
        },
      });
      const findingId = response.createFinding?.findingEdge?.node.id;
      if (findingId != null) {
        void navigate({
          pathname: findingDetailsPath(organizationId, findingId),
          search: location.search,
        });
      }
    } catch {
      // The mutation hook already reported the failure.
    }
  }

  return (
    <form onSubmit={(event) => {
      void submit(event);
    }}
    >
      <TaskDetails
        back={(
          <Link
            to={{ pathname: findingListPath(organizationId), search: location.search }}
            size={2}
            color="neutral"
            underline={false}
            iconStart={<CaretLeftIcon />}
            className={back()}
          >
            {t("details.back")}
          </Link>
        )}
        name={(
          <input
            name="title"
            className={titleInput()}
            required
            disabled={isCreating}
            onChange={(event) => {
              const input = event.currentTarget;
              const clamped = clampFindingTitle(input.value);
              if (clamped !== input.value) {
                input.value = clamped;
              }
            }}
            placeholder={t("create.placeholders.title")}
            aria-label={t("details.fields.title")}
          />
        )}
        actions={(
          <Button
            type="submit"
            variant="solid"
            color="neutral"
            highContrast
            loading={isCreating}
          >
            {t("create.actions.create")}
          </Button>
        )}
        engagement={(
          <div className={narrative()}>
            <Field className={narrativeWide()} label={t("details.fields.source")}>
              <TextField name="source" disabled={isCreating} />
            </Field>
            <Field label={t("details.fields.rootCause")}>
              <Textarea name="rootCause" rows={4} disabled={isCreating} />
            </Field>
            <Field label={t("details.fields.correctiveAction")}>
              <Textarea name="correctiveAction" rows={4} disabled={isCreating} />
            </Field>
            <Field className={narrativeWide()} label={t("details.fields.effectivenessCheck")}>
              <Textarea name="effectivenessCheck" rows={4} disabled={isCreating} />
            </Field>
            <Field className={narrativeWide()} label={t("details.fields.description")}>
              <Textarea name="description" rows={4} disabled={isCreating} />
            </Field>
          </div>
        )}
        properties={(
          <FindingPropertiesLayout>
            <FindingPropertiesKind
              kind={kind}
              canUpdate
              disabled={isCreating}
              onChange={setKind}
            />
            <FindingPropertiesStatus
              status={status}
              canUpdate
              disabled={isCreating}
              onChange={(next) => {
                setStatus(next);
                if (next !== "RISK_ACCEPTED") {
                  setRiskError(null);
                }
              }}
            />
            {status === "RISK_ACCEPTED" && (
              <FindingPropertiesRisk
                risk={risk}
                disabled={isCreating}
                error={riskError}
                onSelect={(next) => {
                  setRisk(next);
                  setRiskError(null);
                }}
              />
            )}
            <FindingPropertiesPriority
              priority={priority}
              canUpdate
              disabled={isCreating}
              onChange={setPriority}
            />
            <FindingPropertiesOwner
              owner={owner}
              canUpdate
              clearable
              disabled={isCreating}
              onChange={(ownerId) => {
                if (ownerId == null) {
                  setOwner(null);
                }
              }}
              onSelect={setOwner}
            />
            <FindingPropertiesDate
              identified
              value={identifiedOn}
              placeholder={t("details.placeholders.dateIdentified")}
              canUpdate
              disabled={isCreating}
              onChange={setIdentifiedOn}
            />
            <FindingPropertiesDate
              value={dueDate}
              placeholder={t("details.placeholders.dueDate")}
              overdue={findingDueIsOverdue(dueDate, status)}
              canUpdate
              disabled={isCreating}
              onChange={setDueDate}
            />
          </FindingPropertiesLayout>
        )}
      />
    </form>
  );
}
