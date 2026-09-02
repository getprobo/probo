// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { todayAsDateInput } from "@probo/helpers";
import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";

import { matrixAsOfParam, parseAsOfDate } from "./matrixAsOf";

function writeAsOfDate(
  params: URLSearchParams,
  dateInput: string,
  options?: { clearCellFilter?: boolean },
): URLSearchParams {
  const next = new URLSearchParams(params);
  const today = todayAsDateInput();
  if (!dateInput || dateInput >= today) {
    next.delete(matrixAsOfParam);
  } else {
    next.set(matrixAsOfParam, dateInput);
  }

  if (options?.clearCellFilter !== false) {
    next.delete("score");
    next.delete("likelihood");
    next.delete("impact");
  }

  return next;
}

export function useMatrixAsOf() {
  const [params, setParams] = useSearchParams();
  const search = params.toString();
  const asOfDate = useMemo(
    () => parseAsOfDate(new URLSearchParams(search)),
    [search],
  );

  const setAsOfDate = useCallback((
    next: string,
    options?: { clearCellFilter?: boolean },
  ) => {
    setParams(prev => writeAsOfDate(prev, next, options), {
      replace: true,
      preventScrollReset: true,
    });
  }, [setParams]);

  return {
    asOfDate,
    setAsOfDate,
  };
}
