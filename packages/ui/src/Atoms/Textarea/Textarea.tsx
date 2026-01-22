import {
  useRef,
  type TextareaHTMLAttributes,
  type RefCallback,
  useLayoutEffect,
  useCallback,
} from "react";
import { input } from "../Input/Input";
import clsx from "clsx";

type Props = TextareaHTMLAttributes<HTMLTextAreaElement> & {
  variant?: "bordered" | "ghost" | "title";
  autogrow?: boolean;
  ref?: RefCallback<HTMLTextAreaElement>;
};

export function Textarea(props: Props) {
  const ref = useRef<HTMLTextAreaElement>(null);
  const { autogrow, ref: propsRef, ...restProps } = props;

  const adjustHeight = useCallback(() => {
    if (!autogrow || !ref.current) return;
    ref.current.style.height = "inherit";
    const paddingY = 2;
    ref.current.style.height = `${ref.current.scrollHeight + paddingY * 2}px`;
  }, [autogrow, ref]);

  useLayoutEffect(() => {
    adjustHeight();
  }, [adjustHeight]);

  return (
    <textarea
      {...restProps}
      ref={(node) => {
        ref.current = node;
        propsRef?.(node);
      }}
      onInput={(e) => {
        adjustHeight();
        props.onInput?.(e);
      }}
      className={input({
        ...props,
        className: clsx("min-h-20", props.className),
      })}
    />
  );
}
