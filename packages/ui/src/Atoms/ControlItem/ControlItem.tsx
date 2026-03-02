import { type HTMLAttributes, useEffect, useRef } from "react";
import { Link } from "react-router";
import { tv } from "tailwind-variants";

type Props = {
  active?: boolean;
  id: string;
  description?: string;
  to: string;
} & HTMLAttributes<HTMLAnchorElement>;

const classNames = tv({
  slots: {
    wrapper: "block p-4 space-y-[6px] rounded-xl cursor-pointer text-start",
    id: "px-[6px] py-[2px] text-base font-medium border border-border-low rounded-lg w-max",
    description: "text-sm text-txt-tertiary line-clamp-3",
  },
  variants: {
    active: {
      true: {
        wrapper: "bg-tertiary-pressed",
        id: "bg-active",
      },
      false: {
        wrapper: "hover:bg-tertiary-hover",
        id: "bg-highlight",
      },
    },
  },
  defaultVariants: {
    active: false,
  },
});

export function ControlItem({ active, id, description, to, ...props }: Props) {
  const {
    wrapper,
    id: idCls,
    description: descriptionCls,
  } = classNames({
    active,
  });

  const ref = useRef<HTMLAnchorElement>(null);

  // Make the active element scroll into view when selected
  useEffect(() => {
    if (ref.current && active) {
      ref.current.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }, [active]);

  return (
    <Link className={wrapper()} to={to} {...props} ref={ref}>
      <div className={idCls()}>{id}</div>
      <div className={descriptionCls()}>{description}</div>
    </Link>
  );
}
