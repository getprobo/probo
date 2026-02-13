import { z } from "zod";

import { useFormWithSchema } from "../useFormWithSchema";

export const documentSchema = z.object({
  title: z.string().min(1, "Title is required"),
  content: z.string().min(1, "Content is required"),
  approverId: z.string().min(1, "Approver is required"),
  documentType: z.enum(["OTHER", "ISMS", "POLICY", "PROCEDURE"]),
  classification: z.enum(["PUBLIC", "INTERNAL", "CONFIDENTIAL", "SECRET"]),
});

export const useDocumentForm = () => {
  return useFormWithSchema(documentSchema, {
    defaultValues: {
      documentType: "POLICY",
      classification: "INTERNAL",
    },
  });
};
