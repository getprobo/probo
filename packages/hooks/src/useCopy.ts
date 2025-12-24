import { useRef, useState } from "react";

export function useCopy(): [boolean, (value: string) => void] {
  const [copiedValue, setCopiedValue] = useState<string>();
  const lastCopiedValueRef = useRef<string>("");

  const handleCopy = (value: string) => {
    lastCopiedValueRef.current = value;
    navigator.clipboard.writeText(value);
    setCopiedValue(value);
    setTimeout(() => {
      setCopiedValue(undefined);
      lastCopiedValueRef.current = "";
    }, 2000);
  };

  return [copiedValue === lastCopiedValueRef.current, handleCopy];
}
