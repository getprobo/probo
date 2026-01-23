import { clsx } from "clsx";

import { Avatar } from "../Avatar/Avatar";

interface FrameworkLogoProps {
  className?: string;
  darkLogoURL?: string | null;
  lightLogoURL?: string | null;
  name: string;
}

export function FrameworkLogo(props: FrameworkLogoProps) {
  const { className, lightLogoURL, darkLogoURL, name } = props;

  return lightLogoURL || darkLogoURL
    ? (
        <>
          {lightLogoURL && (
            <img
              src={lightLogoURL}
              className={clsx("size-12 block dark:hidden", className)}
              alt={name}
            />
          )}
          {darkLogoURL && (
            <img
              src={darkLogoURL}
              className={clsx("size-12 hidden dark:block", className)}
              alt={name}
            />
          )}
        </>
      )
    : (
        <Avatar name={name} size="l" className={clsx("size-12", className)} />
      );
}
