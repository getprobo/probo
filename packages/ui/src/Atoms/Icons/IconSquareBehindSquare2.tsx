import type { IconProps } from "./type";

export function IconSquareBehindSquare2({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M11.25 4C10.5597 4 10 4.55966 10 5.25V8H12.75C14.5449 8 16 9.45512 16 11.25V14H18.75C19.4403 14 20 13.4403 20 12.75V5.25C20 4.55966 19.4403 4 18.75 4H11.25ZM16 16H18.75C20.5449 16 22 14.5449 22 12.75V5.25C22 3.45506 20.5449 2 18.75 2H11.25C9.45513 2 8 3.45506 8 5.25V8H5.25C3.45506 8 2 9.45513 2 11.25V18.75C2 20.5449 3.45506 22 5.25 22H12.75C14.5449 22 16 20.5449 16 18.75V16ZM14 11.25C14 10.5597 13.4403 10 12.75 10H5.25C4.55966 10 4 10.5597 4 11.25V18.75C4 19.4403 4.55966 20 5.25 20H12.75C13.4403 20 14 19.4403 14 18.75V11.25Z" fill="currentColor" />
    </svg>
  );
}
