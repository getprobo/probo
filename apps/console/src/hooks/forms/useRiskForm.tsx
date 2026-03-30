// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type {
  useRiskFormFragment$data,
  useRiskFormFragment$key,
} from "#/__generated__/core/useRiskFormFragment.graphql";

import { useFormWithSchema } from "../useFormWithSchema";

const RiskFragment = graphql`
  fragment useRiskFormFragment on Risk {
    id
    # eslint-disable-next-line relay/unused-fields
    name
    # eslint-disable-next-line relay/unused-fields
    category
    description
    # eslint-disable-next-line relay/unused-fields
    treatment
    # eslint-disable-next-line relay/unused-fields
    inherentLikelihood
    # eslint-disable-next-line relay/unused-fields
    inherentImpact
    # eslint-disable-next-line relay/unused-fields
    residualLikelihood
    # eslint-disable-next-line relay/unused-fields
    residualImpact
    # eslint-disable-next-line relay/unused-fields
    note
    owner {
      id
    }
  }
`;

export type RiskNode = useRiskFormFragment$data;
export type RiskKey = useRiskFormFragment$key & { id: string };

// Export the schema so it can be used elsewhere
export const riskSchema = z.object({
  category: z.string().min(1, "Category is required"),
  name: z.string().min(1, "Name is required"),
  description: z.string().optional().nullable(),
  ownerId: z.string().min(1, "Owner is required"),
  treatment: z.enum(["AVOIDED", "MITIGATED", "TRANSFERRED", "ACCEPTED"]),
  inherentLikelihood: z.number({ coerce: true }).min(1).max(5),
  inherentImpact: z.number({ coerce: true }).min(1).max(5),
  residualLikelihood: z.number({ coerce: true }).min(1).max(5),
  residualImpact: z.number({ coerce: true }).min(1).max(5),
  note: z.string().optional(),
});

export const useRiskForm = (riskKey?: RiskKey) => {
  const risk = useFragment(RiskFragment, riskKey);
  return useFormWithSchema(riskSchema, {
    defaultValues: risk
      ? {
          ...risk,
          description: risk.description ?? undefined,
          ownerId: risk.owner?.id,
        }
      : {
          inherentLikelihood: 3,
          inherentImpact: 3,
          residualLikelihood: 3,
          residualImpact: 3,
        },
  });
};

export type RiskForm = ReturnType<typeof useRiskForm>;

export type RiskData = z.infer<typeof riskSchema>;
