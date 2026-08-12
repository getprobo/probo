import type { ComponentProps } from "react";

export function Front(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <rect width="24" height="24" rx="5" fill="#001B38" />
      <path
        fill="#FFFFFF"
        d="M8 6h8.5v2.6h-5.9v2.3h4.6v2.6h-4.6V18H8V6z"
      />
    </svg>
  );
}
