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

import { useRefSync } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { type ReactNode, useEffect } from "react";
import { useInView } from "react-intersection-observer";

import { Spinner } from "../Spinner/Spinner";

type Props = {
  children?: ReactNode;
  onView: () => void;
  loading?: boolean;
};

export function InfiniteScrollTrigger({ children, onView, loading }: Props) {
  const { __ } = useTranslate();
  const { ref, inView } = useInView({
    threshold: 0,
  });
  const onViewRef = useRefSync(onView);
  useEffect(() => {
    if (inView && !loading) onViewRef.current();
  }, [inView, loading, onViewRef]);

  return (
    <div
      className="flex gap-2 items-center justify-center text-xs text-txt-secondary"
      ref={ref}
    >
      {children ?? (
        <>
          <Spinner size={16} />
          {__("Loading")}
        </>
      )}
    </div>
  );
}
