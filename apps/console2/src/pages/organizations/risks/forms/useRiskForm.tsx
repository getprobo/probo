import { z } from "zod";
import { useFormWithSchema } from "../../../../hooks/useFormWithSchema";

// Export the schema so it can be used elsewhere
export const riskSchema = z.object({
  category: z.string(),
  name: z.string(),
  description: z.string(),
  organizationId: z.string(),
  ownerId: z.string(),
  treatment: z.enum(["AVOIDED", "MITIGATED", "TRANSFERRED", "ACCEPTED"]),
  inherentLikelihood: z.number().min(1).max(5),
  inherentImpact: z.number().min(1).max(5),
  residualLikelihood: z.number().min(1).max(5),
  residualImpact: z.number().min(1).max(5),
  note: z.string(),
});

export const useRiskForm = (organizationId: string) => {
  return useFormWithSchema(riskSchema, {
    defaultValues: {
      category: "Compliance & Legal",
      name: "Contractual risk due to poorly drafted agreements",
      description:
        "Weak contracts leads to disputes, missed deliverables, or revenue leakage",
      organizationId: organizationId,
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
