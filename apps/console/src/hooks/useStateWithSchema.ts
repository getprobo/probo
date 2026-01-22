import { useCallback, useMemo, useState } from "react";
import { z, ZodError, type ZodTypeAny } from "zod";

export function useStateWithSchema<T extends ZodTypeAny>(
  schema: T,
  initialValue: z.infer<T>,
) {
  const [state, setState] = useState(initialValue);
  const [value, errors] = useMemo((): [z.infer<T>, Record<string, string>] => {
    try {
      schema.parse(state);
      return [schema.parse(state) as z.TypeOf<T>, {}];
    } catch (error) {
      if (error instanceof ZodError) {
        return [
          state,
          Object.fromEntries(
            error.issues.map(issue => [
              issue.path.join("."),
              issue.message,
            ]) ?? [],
          ),
        ];
      }
      return [state, {}];
    }
  }, [state, schema]);

  return {
    rawValue: state,
    value,
    errors,
    update: useCallback(
      <TKey extends keyof z.infer<T>>(key: TKey, value: z.infer<T>[TKey]) => {
        setState(prevState => ({ ...prevState, [key]: value }));
      },
      [],
    ),
  };
}
