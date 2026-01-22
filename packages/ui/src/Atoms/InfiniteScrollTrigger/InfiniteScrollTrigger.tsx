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
