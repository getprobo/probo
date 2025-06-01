import { z } from "zod";
import { useFormWithSchema } from "../useFormWithSchema";

export const documentSchema = z.object({
  title: z.string(),
  content: z.string(),
  ownerId: z.string(),
});

export const useDocumentForm = () => {
  return useFormWithSchema(documentSchema, {});
};
