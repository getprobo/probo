import { z } from "zod";
import { useFormWithSchema } from "../useFormWithSchema";

export const documentSchema = z.object({
  title: z.string().min(1, "Title is required"),
  content: z.string().min(1, "Content is required"),
  ownerId: z.string().min(1, "Owner is required"),
  documentType: z.enum(["OTHER", "ISMS", "POLICY"]),
});

export const useDocumentForm = () => {
  return useFormWithSchema(documentSchema, {
    defaultValues: {
      documentType: "POLICY",
    },
  });
};
