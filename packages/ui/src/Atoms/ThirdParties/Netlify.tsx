import type { ComponentProps } from "react";

export function Netlify(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <path
        d="M6.49 19.04h-.23L5.13 17.91v-.23l3.46-3.46h1.92l.25.25v1.92zM5.13 6.31V6.08l1.13-1.13h.23l3.46 3.46v1.92l-.25.25H7.78zm9.96 9.09h-2.65l-.22-.22V8.83c0-.39-.15-.74-.5-.85-.36-.11-.74.03-.95.34l-.27.27h-1.93l-.25-.25V6.41l.22-.22h6.55l.22.22v8.77zm0-13.78v2.21l-.22.22h-1.49l-.22-.22V1.62l.22-.22h1.49zm0 17.55v2.21l-.22.22h-1.49l-.22-.22v-2.21l.22-.22h1.49zM20.5 12.93h-2.21l-.22-.22v-1.49l.22-.22h2.21l.22.22v1.49zm-13.34 0H1.5l-.22-.22v-1.49l.22-.22h5.66l.22.22v1.49z"
        fill="#00C7B7"
      />
    </svg>
  );
}
