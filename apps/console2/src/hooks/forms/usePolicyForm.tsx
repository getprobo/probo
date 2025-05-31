import { z } from "zod";
import { useFormWithSchema } from "../useFormWithSchema";

export const policySchema = z.object({
  title: z.string(),
  content: z.string(),
  ownerId: z.string(),
});

export const usePolicyForm = () => {
  return useFormWithSchema(policySchema, {});
};
