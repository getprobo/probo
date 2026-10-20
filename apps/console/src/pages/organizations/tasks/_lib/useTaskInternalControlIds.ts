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
import { graphql, usePaginationFragment } from "react-relay";

import type { useTaskInternalControlIds_task$key } from "#/__generated__/core/useTaskInternalControlIds_task.graphql";
import type { useTaskInternalControlIdsQuery } from "#/__generated__/core/useTaskInternalControlIdsQuery.graphql";

const pageSize = 50;

const taskInternalControlIdsFragment = graphql`
  fragment useTaskInternalControlIds_task on Task
  @refetchable(queryName: "useTaskInternalControlIdsQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey" }
  ) {
    id
    internalControls(first: $first, after: $after)
      @connection(key: "useTaskInternalControlIds_internalControls") {
      edges {
        node {
          id
        }
      }
    }
  }
`;

export function useTaskInternalControlIds(taskKey: useTaskInternalControlIds_task$key) {
  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment<
    useTaskInternalControlIdsQuery,
    useTaskInternalControlIds_task$key
  >(taskInternalControlIdsFragment, taskKey);
  const [failed, setFailed] = useState(false);
  const retry = useCallback(() => {
    setFailed(false);
  }, []);

  useEffect(() => {
    if (!hasNext || isLoadingNext || failed) {
      return;
    }

    loadNext(pageSize, {
      onComplete: (error) => {
        if (error) {
          setFailed(true);
        }
      },
    });
  }, [failed, hasNext, isLoadingNext, loadNext]);

  return {
    taskId: data.id,
    ids: data.internalControls.edges.map(edge => edge.node.id),
    pending: !failed && (hasNext || isLoadingNext),
    failed,
    retry,
  };
}
