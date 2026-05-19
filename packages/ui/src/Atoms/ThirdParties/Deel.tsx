import type { ComponentProps } from "react";

export function Deel(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <path
        d="M3 3h7.5c4.97 0 9 4.03 9 9s-4.03 9-9 9H3V3zm3.5 3.2v11.6h4c3.2 0 5.8-2.6 5.8-5.8S13.7 6.2 10.5 6.2h-4z"
        fill="#15847B"
      />
    </svg>
  );
}
