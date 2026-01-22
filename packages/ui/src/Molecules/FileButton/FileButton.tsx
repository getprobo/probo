import type {
  ChangeEventHandler,
  ComponentProps,
  PropsWithChildren,
  RefObject,
} from "react";

import { button, Button } from "../../Atoms/Button/Button";

type Props = PropsWithChildren<{
  onChange: ChangeEventHandler<HTMLInputElement>;
  accept?: string;
  disabled?: boolean;
  className?: string;
  ref?: RefObject<HTMLInputElement | null>;
}>
& Pick<ComponentProps<typeof Button>, "disabled" | "variant" | "icon">;

export function FileButton({
  onChange,
  children,
  icon: IconComponent,
  ref,
  accept,
  ...props
}: Props) {
  return (
    <label className={button({ ...props })}>
      {IconComponent && <IconComponent size={16} className="flex-none" />}
      {children}
      <input
        type="file"
        onChange={onChange}
        hidden
        ref={ref}
        {...(accept && { accept })}
      />
    </label>
  );
}
