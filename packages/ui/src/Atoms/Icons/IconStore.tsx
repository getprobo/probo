import type { IconProps } from "./type";

export function IconStore({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M5 12V20C5 20.5527 5.44728 21 6 21H18C18.5527 21 19 20.5527 19 20V12H21V20C21 21.6573 19.6573 23 18 23H6C4.34271 23 3 21.6573 3 20V12H5Z" fill="currentColor" />
      <path fillRule="evenodd" clipRule="evenodd" d="M9 15H15V22H13V17H11V22H9V15Z" fill="currentColor" />
      <path fillRule="evenodd" clipRule="evenodd" d="M4.42137 1H19.5786L22.944 6.8548L22.8424 7.24927C22.2875 9.40529 20.3314 11 18 11C16.864 11 15.8294 10.6148 15 9.98264C14.1706 10.6148 13.136 11 12 11C10.864 11 9.82943 10.6148 9 9.98264C8.17057 10.6148 7.136 11 6 11C3.66861 11 1.7125 9.40529 1.15756 7.24927L1.05603 6.8548L4.42137 1ZM5.57862 3L3.21355 7.1146C3.65602 8.21984 4.73789 9 6 9C6.89162 9 7.68903 8.60064 8.24692 7.96207L9 7.10009L9.75307 7.96207C10.311 8.60064 11.1084 9 12 9C12.8916 9 13.689 8.60064 14.2469 7.96207L15 7.10009L15.7531 7.96207C16.311 8.60064 17.1084 9 18 9C19.2621 9 20.344 8.21984 20.7864 7.1146L18.4214 3H5.57862Z" fill="currentColor" />
    </svg>
  );
}
