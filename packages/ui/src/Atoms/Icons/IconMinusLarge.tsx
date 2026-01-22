import type { IconProps } from "./type";

export function IconMinusLarge({ size = 16, className }: IconProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 16 16"
      fill="currentColor"
      className={className}
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M13.3333 7.33203C13.7015 7.33203 14 7.63051 14 7.9987C14 8.36689 13.7015 8.66536 13.3333 8.66536H2.66667C2.29848 8.66536 2 8.36689 2 7.9987C2 7.63051 2.29848 7.33203 2.66667 7.33203H13.3333Z"
        fill="currentColor"
      />
    </svg>
  );
}
