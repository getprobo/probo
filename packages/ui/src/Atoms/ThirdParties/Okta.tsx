import type { ComponentProps } from "react";

export function Okta(props: ComponentProps<"svg">) {
  return (
    <svg
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M12 2C6.477 2 2 6.477 2 12s4.477 10 10 10 10-4.477 10-10S17.523 2 12 2zm0 5a5 5 0 100 10 5 5 0 000-10z"
        fill="#007DC1"
      />
    </svg>
  );
}
