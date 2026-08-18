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

import { useCallback } from "react";
import { useSearchParams } from "react-router";

import {
  isMatrixScoreType,
  type MatrixCell,
  sameMatrixCell,
} from "./matrixCell";

function parseAxis(value: string | null): number | null {
  if (value == null || value === "") {
    return null;
  }

  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed < 1) {
    return null;
  }

  return parsed;
}

function parseCell(params: URLSearchParams): MatrixCell | null {
  const type = params.get("score");
  const likelihood = parseAxis(params.get("likelihood"));
  const impact = parseAxis(params.get("impact"));
  if (!isMatrixScoreType(type) || likelihood == null || impact == null) {
    return null;
  }

  return { type, likelihood, impact };
}

function writeCell(params: URLSearchParams, cell: MatrixCell | null): URLSearchParams {
  const next = new URLSearchParams(params);
  if (cell == null) {
    next.delete("score");
    next.delete("likelihood");
    next.delete("impact");
    return next;
  }

  next.set("score", cell.type);
  next.set("likelihood", String(cell.likelihood));
  next.set("impact", String(cell.impact));
  return next;
}

export function useMatrixCellFilter() {
  const [params, setParams] = useSearchParams();
  const cell = parseCell(params);

  const setCell = useCallback((next: MatrixCell | null) => {
    setParams(prev => writeCell(prev, next), { replace: true });
  }, [setParams]);

  const toggleCell = useCallback((next: MatrixCell) => {
    if (cell && sameMatrixCell(cell, next)) {
      setCell(null);
      return;
    }

    setCell(next);
  }, [cell, setCell]);

  return {
    cell,
    toggleCell,
    clear: () => setCell(null),
  };
}
