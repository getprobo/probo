import { z } from "zod";
import { useFormWithSchema } from "../../../../hooks/useFormWithSchema";
import { graphql } from "relay-runtime";
import type {
  useRiskFormFragment$data,
  useRiskFormFragment$key,
} from "./__generated__/useRiskFormFragment.graphql";
import { useFragment } from "react-relay";

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
  category: z.string(),
  name: z.string(),
  description: z.string(),
  ownerId: z.string(),
  treatment: z.enum([
    "AVOIDED",
    "MITIGATED",
    "TRANSFERRED",
    "ACCEPTED",
    "%future added value",
  ]),
  inherentLikelihood: z.number({ coerce: true }).min(1).max(5),
  inherentImpact: z.number({ coerce: true }).min(1).max(5),
  residualLikelihood: z.number({ coerce: true }).min(1).max(5),
  residualImpact: z.number({ coerce: true }).min(1).max(5),
  note: z.string(),
});

export const useRiskForm = (riskKey?: RiskKey) => {
  const risk = useFragment(RiskFragment, riskKey);
  return useFormWithSchema(riskSchema, {
    defaultValues: risk
      ? {
          ...risk,
          ownerId: risk.owner?.id,
        }
      : {
          category: "Compliance & Legal",
          name: "Contractual risk due to poorly drafted agreements",
          description:
            "Weak contracts leads to disputes, missed deliverables, or revenue leakage",
          ownerId: "a_YJLJ5RAAIACAAAAZbuSmzPhFPhoMEQ",
          treatment: "MITIGATED",
          inherentLikelihood: 5,
          inherentImpact: 5,
          residualLikelihood: 5,
          residualImpact: 5,
          note: "Hello worlds this is a test",
        },
  });
};

export type RiskForm = ReturnType<typeof useRiskForm>;

export type RiskData = z.infer<typeof riskSchema>;
