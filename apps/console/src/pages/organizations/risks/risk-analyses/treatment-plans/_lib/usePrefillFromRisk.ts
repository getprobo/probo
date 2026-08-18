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

import { useEffect } from "react";
import { fetchQuery, graphql, useRelayEnvironment } from "react-relay";

import type {
  RiskTreatment,
  usePrefillFromRiskQuery,
} from "#/__generated__/core/usePrefillFromRiskQuery.graphql";

const prefillFromRiskQuery = graphql`
  query usePrefillFromRiskQuery($riskId: ID!) {
    node(id: $riskId) {
      __typename
      ... on Risk {
        id
        treatment
        owner {
          id
        }
      }
    }
  }
`;

export type TreatmentPlanPrefillValues = {
  riskId: string;
  treatment: RiskTreatment;
  ownerId: string;
};

function emptyPrefill(riskId: string): TreatmentPlanPrefillValues {
  return {
    riskId,
    treatment: "MITIGATED",
    ownerId: "",
  };
}

export function usePrefillFromRisk(
  riskId: string,
  enabled: boolean,
  onPrefill: (values: TreatmentPlanPrefillValues) => void,
) {
  const environment = useRelayEnvironment();

  useEffect(() => {
    if (!enabled || !riskId) {
      return;
    }

    let disposed = false;
    const subscription = fetchQuery<usePrefillFromRiskQuery>(
      environment,
      prefillFromRiskQuery,
      { riskId },
    ).subscribe({
      next(data) {
        if (disposed) {
          return;
        }

        const risk = data.node?.__typename === "Risk" ? data.node : null;
        if (risk?.id !== riskId) {
          onPrefill(emptyPrefill(riskId));
          return;
        }

        onPrefill({
          riskId,
          treatment: risk.treatment ?? "MITIGATED",
          ownerId: risk.owner?.id ?? "",
        });
      },
      error() {
        if (disposed) {
          return;
        }

        onPrefill(emptyPrefill(riskId));
      },
    });

    return () => {
      disposed = true;
      subscription.unsubscribe();
    };
  }, [enabled, environment, onPrefill, riskId]);
}
