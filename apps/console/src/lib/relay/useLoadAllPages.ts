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

import { useCallback, useEffect, useState } from "react";
import type { LoadMoreFn } from "react-relay";
import type { OperationType } from "relay-runtime";

const maxAttempts = 3;

interface LoadAllPagesOptions<TQuery extends OperationType> {
  hasNext: boolean;
  isLoadingNext: boolean;
  loadNext: LoadMoreFn<TQuery>;
  pageSize: number;
}

// Loads every remaining page of a pagination fragment in the background. A
// failed page is retried with backoff, then the loop stops until `retry`.
export function useLoadAllPages<TQuery extends OperationType>({
  hasNext,
  isLoadingNext,
  loadNext,
  pageSize,
}: LoadAllPagesOptions<TQuery>) {
  const [failures, setFailures] = useState(0);

  useEffect(() => {
    if (!hasNext || isLoadingNext || failures >= maxAttempts) {
      return;
    }
    const timeout = setTimeout(() => {
      loadNext(pageSize, {
        onComplete: (error) => {
          setFailures(count => (error ? count + 1 : 0));
        },
      });
    }, failures === 0 ? 0 : 1000 * 2 ** (failures - 1));
    return () => clearTimeout(timeout);
  }, [failures, hasNext, isLoadingNext, loadNext, pageSize]);

  const retry = useCallback(() => setFailures(0), []);

  return { retry };
}
