import { z } from "zod";
import { useFormWithSchema } from "../useFormWithSchema";

export const documentSchema = z.object({
  title: z.string(),
  content: z.string(),
  ownerId: z.string(),
  documentType: z.enum(["OTHER", "ISMS", "POLICY", "PROCEDURE"]),
});

export const useDocumentForm = () => {
  return useFormWithSchema(documentSchema, {
    defaultValues: {
      documentType: "POLICY",
    },
  });
};
