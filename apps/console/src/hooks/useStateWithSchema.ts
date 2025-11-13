import { z, ZodError, type ZodTypeAny } from "zod";
import { useMemo, useState } from "react";

export function useStateWithSchema<T extends ZodTypeAny>(
  schema: T,
  initialValue: z.infer<T>,
) {
  const [state, setState] = useState(initialValue);
  const errors = useMemo(() => {
    try {
      schema.parse(state);
      return {};
    } catch (error) {
      if (error instanceof ZodError) {
        return Object.fromEntries(
          error.issues.map((issue) => [issue.path.join("."), issue.message]) ??
            [],
        );
      }
      return {};
    }
  }, [state, schema]);

  return [
    state,
    (key: keyof z.infer<T>, value: z.infer<T>[typeof key]) => {
      setState((prevState) => ({ ...prevState, [key]: value }));
    },
    errors,
  ] as const;
}
