import type { IconProps } from "./type";

export function IconArrowUp({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M12 3C12.3978 3 12.7793 3.15803 13.0606 3.43934L18.5607 8.93934C19.1464 9.52512 19.1464 10.4749 18.5607 11.0607C17.9749 11.6464 17.0251 11.6464 16.4393 11.0607L13.5 8.12131V19.5C13.5 20.3284 12.8284 21 12 21C11.1716 21 10.5 20.3284 10.5 19.5V8.12133L7.56066 11.0607C6.97488 11.6464 6.02513 11.6464 5.43934 11.0607C4.85355 10.4749 4.85355 9.52513 5.43934 8.93934L10.9393 3.43934C11.2206 3.15804 11.6022 3 12 3Z" fill="currentColor" />
    </svg>
  );
}
