import type { ComponentProps } from "react";

export function UniFi(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <path
        fill="#0559C9"
        d="M3.6 2.4h3.6v3.6H3.6zM3.6 8.4h3.6v4.8a4.8 4.8 0 009.6 0V8.4h3.6v4.8a8.4 8.4 0 01-16.8 0z"
      />
    </svg>
  );
}
