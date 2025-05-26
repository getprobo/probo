import {
    useRef,
    type FormEventHandler,
    type TextareaHTMLAttributes,
    type RefObject,
} from "react";
import { input } from "../Input/Input";
import clsx from "clsx";

type Props = TextareaHTMLAttributes<HTMLTextAreaElement> & {
    variant?: "bordered" | "ghost" | "title";
    autogrow?: boolean;
    ref?: RefObject<HTMLTextAreaElement>;
};

export function Textarea(props: Props) {
    const textbox = useRef<HTMLTextAreaElement>(null);
    const { autogrow, variant, ref: propsRef, ...restProps } = props;
    const ref = propsRef || textbox;

    const adjustHeight: FormEventHandler<HTMLTextAreaElement> = (e) => {
        props.onInput?.(e);
        if (!autogrow || !ref.current) return;
        ref.current.style.height = "inherit";
        const paddingY = 2;
        ref.current.style.height = `${ref.current.scrollHeight + paddingY * 2}px`;
    };
    return (
        <textarea
            {...restProps}
            ref={ref}
            onInput={adjustHeight}
            className={input({
                ...props,
                className: clsx("min-h-20", props.className),
            })}
        />
    );
}
