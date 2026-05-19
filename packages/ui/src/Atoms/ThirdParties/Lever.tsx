import type { ComponentProps } from "react";

export function Lever(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <path
        d="M3 3h3.6v14.8H21V21H3V3zm6.4 0h3.6v11.2h7V17.5H9.4V3z"
        fill="#16323F"
      />
    </svg>
  );
}
