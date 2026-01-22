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
    name
    category
    description
    treatment
    inherentLikelihood
    inherentImpact
    residualLikelihood
    residualImpact
    inherentRiskScore
    residualRiskScore
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
