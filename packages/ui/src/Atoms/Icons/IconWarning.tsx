import type { IconProps } from "./type";

export function IconWarning({ size = 24, className }: IconProps) {
  return <svg className={className} xmlns="http://www.w3.org/2000/svg" width={size} height={size} fill="currentColor" viewBox="0 0 24 24"><path fill="currentColor" d="M12.188 1.993a3.001 3.001 0 0 1 2.41 1.51L22.596 17.5l.093.174A3 3 0 0 1 20.001 22H4v-.001a2.999 2.999 0 0 1-2.618-4.495l8-14a3 3 0 0 1 2.608-1.518l.198.007Zm-.33 2.002a1 1 0 0 0-.738.498l-.001.003-8 14-.003.004A1.002 1.002 0 0 0 3.99 20h16.008a1.001 1.001 0 0 0 .999-1 1 1 0 0 0-.134-.5l-.002-.004-8-14-.002-.003a1 1 0 0 0-.87-.507l-.132.009ZM12.01 16a1 1 0 0 1 0 2H12a1 1 0 0 1 0-2h.01ZM11 13V9a1 1 0 0 1 2 0v4a1 1 0 1 1-2 0Z" /></svg>;
}
