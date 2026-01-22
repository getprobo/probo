import type { IconProps } from "./type";

export function IconArrowLink({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M20.5607 3.43934C21.1464 4.02513 21.1464 4.97487 20.5607 5.56066L5.56066 20.5607C4.97487 21.1464 4.02513 21.1464 3.43934 20.5607C2.85355 19.9749 2.85355 19.0251 3.43934 18.4393L18.4393 3.43934C19.0251 2.85355 19.9749 2.85355 20.5607 3.43934Z" fill="currentColor" />
      <path fillRule="evenodd" clipRule="evenodd" d="M3 4.5C3 3.67157 3.67157 3 4.5 3H19.5C20.3284 3 21 3.67157 21 4.5V19.5C21 20.3284 20.3284 21 19.5 21C18.6716 21 18 20.3284 18 19.5V6H4.5C3.67157 6 3 5.32843 3 4.5Z" fill="currentColor" />
    </svg>
  );
}
