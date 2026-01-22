import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import type { z, ZodTypeAny } from "zod";

export function useFormWithSchema<T extends ZodTypeAny>(
  schema: T,
  options: Parameters<typeof useForm<z.infer<T>>>[0],
) {
  return useForm<z.infer<T>>({
    ...options,
    resolver: zodResolver(schema),
  });
}
