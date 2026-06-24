// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
