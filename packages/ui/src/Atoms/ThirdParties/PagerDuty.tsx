import type { ComponentProps } from "react";

export function PagerDuty(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <path
        d="M5 2h7.5c4.14 0 7.5 3.36 7.5 7.5S16.64 17 12.5 17H8.5v5H5V2zm3.5 3.2v8.6h4c2.37 0 4.3-1.93 4.3-4.3S14.87 5.2 12.5 5.2h-4z"
        fill="#06AC38"
      />
    </svg>
  );
}
