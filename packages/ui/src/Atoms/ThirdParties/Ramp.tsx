import type { ComponentProps } from "react";

export function Ramp(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <path
        d="M3 3h11.5c2.76 0 5 2.24 5 5 0 2.05-1.23 3.81-2.99 4.59L20 21h-3.7l-3.16-7.5H6.5V21H3V3zm3.5 3.2v4.1h7.6c1.13 0 2.05-.92 2.05-2.05S15.23 6.2 14.1 6.2H6.5z"
        fill="#FFCC33"
      />
    </svg>
  );
}
