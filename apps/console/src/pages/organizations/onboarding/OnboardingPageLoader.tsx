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

import { useCallback, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { onboardingIamQueriesOnboardingIamStatusQuery } from "#/__generated__/iam/onboardingIamQueriesOnboardingIamStatusQuery.graphql";
import type { OnboardingPageQuery } from "#/__generated__/core/OnboardingPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import {
  OnboardingPage,
  onboardingPageQuery,
} from "./OnboardingPage";
import { onboardingIamStatusQuery } from "#/pages/iam/organizations/onboarding/onboardingIamQueries";

export default function OnboardingPageLoader() {
  const organizationId = useOrganizationId();
  const [coreQueryRef, loadCoreQuery] = useQueryLoader<OnboardingPageQuery>(
    onboardingPageQuery,
  );
  const [iamQueryRef, loadIamQuery] = useQueryLoader<onboardingIamQueriesOnboardingIamStatusQuery>(
    onboardingIamStatusQuery,
  );

  useEffect(() => {
    loadCoreQuery({ organizationId });
  }, [loadCoreQuery, organizationId]);

  useEffect(() => {
    loadIamQuery({ organizationId });
  }, [loadIamQuery, organizationId]);

  const refetchIamStatus = useCallback(() => {
    loadIamQuery({ organizationId }, { fetchPolicy: "network-only" });
  }, [loadIamQuery, organizationId]);

  if (!coreQueryRef || !iamQueryRef) {
    return null;
  }

  return (
    <IAMRelayProvider>
      <OnboardingPage
        coreQueryRef={coreQueryRef}
        iamQueryRef={iamQueryRef}
        refetchIamStatus={refetchIamStatus}
      />
    </IAMRelayProvider>
  );
}
