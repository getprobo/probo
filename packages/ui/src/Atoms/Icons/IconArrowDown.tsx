import type { IconProps } from "./type";

export function IconArrowDown({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M12 21C11.6022 21 11.2207 20.842 10.9394 20.5607L5.43934 15.0607C4.85355 14.4749 4.85355 13.5251 5.43934 12.9393C6.02512 12.3536 6.97487 12.3536 7.56066 12.9393L10.5 15.8787V4.5C10.5 3.67157 11.1716 3 12 3C12.8284 3 13.5 3.67157 13.5 4.5V15.8787L16.4393 12.9393C17.0251 12.3536 17.9749 12.3536 18.5607 12.9393C19.1464 13.5251 19.1464 14.4749 18.5607 15.0607L13.0607 20.5607C12.7794 20.842 12.3978 21 12 21Z" fill="currentColor" />
    </svg>
  );
}
