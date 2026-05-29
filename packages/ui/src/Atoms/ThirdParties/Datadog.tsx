import type { ComponentProps } from "react";

export function Datadog(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <path
        d="M20.4 3 3.6 8.6l3.5 3.4-2 2.2 2.5 2.6-1.1 3.2 4.4-3.6 5.9-.4L20.4 3zm-7 11.6-2.2-1.4 1.5-2.6 2.5-.6-1.8 4.6z"
        fill="#632CA6"
      />
    </svg>
  );
}
