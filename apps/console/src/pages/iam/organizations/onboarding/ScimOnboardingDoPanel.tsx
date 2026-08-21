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

import { Link } from "@probo/ui";
import { Suspense, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { usePreloadedQuery, useQueryLoader } from "react-relay";

import type { onboardingIamQueriesScimOnboardingDoPanelQuery } from "#/__generated__/iam/onboardingIamQueriesScimOnboardingDoPanelQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";
import {
  OnboardingStepActions,
  OnboardingStepDoCard,
} from "#/pages/organizations/onboarding/_components/OnboardingStepActions";

import { ConnectorList } from "../settings/_components/ConnectorList";
import { SCIMConfiguration } from "../settings/_components/SCIMConfiguration";
import { scimOnboardingDoPanelQuery } from "./onboardingIamQueries";

type ScimOnboardingDoPanelInnerProps = {
  onContinue: () => void;
  onDefer: () => void;
  queryRef: NonNullable<
    ReturnType<
      typeof useQueryLoader<onboardingIamQueriesScimOnboardingDoPanelQuery>
    >[0]
  >;
};

function ScimOnboardingDoPanelInner(props: ScimOnboardingDoPanelInnerProps & {
  onScimUpdated?: () => void;
}) {
  const { onContinue, onDefer, onScimUpdated, queryRef } = props;
  const { t } = useTranslation("organizations/onboarding");
  const organizationId = useOrganizationId();

  const { organization } = usePreloadedQuery<onboardingIamQueriesScimOnboardingDoPanelQuery>(
    scimOnboardingDoPanelQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid organization node");
  }

  const complete = !!organization.scimConfiguration?.id;

  useEffect(() => {
    if (complete) {
      onScimUpdated?.();
    }
  }, [complete, onScimUpdated]);

  const settingsHref = `/organizations/${organizationId}/settings/scim`;

  return (
    <OnboardingStepDoCard
      title={t("steps.scim.doTitle")}
      description={t("steps.scim.doDescription")}
      complete={complete}
      actions={(
        <OnboardingStepActions
          continueDisabled={!complete}
          onContinue={onContinue}
          onDefer={onDefer}
          settingsLink={(
            <Link to={settingsHref} variant="secondary" size="2">
              {t("actions.openInSettings")}
            </Link>
          )}
        />
      )}
    >
      <div className="space-y-6 max-h-[min(28rem,50vh)] overflow-y-auto pr-1">
        <ConnectorList fKey={organization} />
        <div className="space-y-2 border-t border-border-mid pt-4">
          <h4 className="text-2 font-medium">{t("steps.scim.manualScimHeading")}</h4>
          <SCIMConfiguration fKey={organization} />
        </div>
      </div>
    </OnboardingStepDoCard>
  );
}

export function ScimOnboardingDoPanel(props: {
  onContinue: () => void;
  onDefer: () => void;
  onScimUpdated?: () => void;
}) {
  const { onContinue, onDefer, onScimUpdated } = props;
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<onboardingIamQueriesScimOnboardingDoPanelQuery>(
    scimOnboardingDoPanelQuery,
  );

  useEffect(() => {
    loadQuery({ organizationId }, { fetchPolicy: "store-and-network" });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return null;
  }

  return (
    <IAMRelayProvider>
      <Suspense fallback={null}>
        <ScimOnboardingDoPanelInner
          onContinue={onContinue}
          onDefer={onDefer}
          onScimUpdated={onScimUpdated}
          queryRef={queryRef}
        />
      </Suspense>
    </IAMRelayProvider>
  );
}
