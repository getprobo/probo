import { clsx } from "clsx";
import { Fragment } from "react/jsx-runtime";
import { Link } from "react-router";

import { IconChevronRight } from "../Icons";

type Props = {
  items: (ItemProps | string)[];
};

type ItemProps = {
  label: string;
  to?: string;
  active?: boolean;
};

export function Breadcrumb({ items }: Props) {
  return (
    <div className="flex items-center gap-[6px] text-txt-tertiary">
      {items.map((item, k) => (
        <Fragment key={k}>
          {k > 0 && <IconChevronRight size={12} />}
          <BreadcrumbItem
            {...(typeof item === "string" ? { label: item } : item)}
            active={k === items.length - 1}
          />
        </Fragment>
      ))}
    </div>
  );
}

function BreadcrumbItem({ label, to, active }: ItemProps) {
  const className = clsx(
    "text-sm px-1 rounded-sm h-5",
    active && "text-txt-primary font-medium",
  );
  if (to) {
    return (
      <Link
        to={to}
        className={clsx(
          className,
          "hover:bg-tertiary-hover active:bg-tertiary-pressed",
        )}
      >
        {label}
      </Link>
    );
  }
  return <span className={className}>{label}</span>;
}
