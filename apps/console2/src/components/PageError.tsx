import { useLocation, useRouteError } from "react-router";
import { IconPageCross } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useEffect, useRef } from "react";

const classNames = {
  wrapper: "py-10 text-center space-y-2 ",
  title: "text-2xl flex gap-2 font-semibold items-center justify-center",
  description: "text-base text-txt-tertiary",
};

type Props = {
  resetErrorBoundary: () => void;
};

export function PageError({ resetErrorBoundary }: Props) {
  const error = useRouteError();
  const { __ } = useTranslate();
  const location = useLocation();
  const baseLocation = useRef(location);

  // Reset error boundary on page change
  useEffect(() => {
    if (location.pathname !== baseLocation.current.pathname) {
      resetErrorBoundary();
    }
  }, [location, resetErrorBoundary]);

  if (!error) {
    return (
      <div className={classNames.wrapper}>
        <h1 className={classNames.title}>
          <IconPageCross size={26} />
          {__("Page not found")}
        </h1>
        <p className={classNames.description}>
          {__("The page you are looking for does not exist")}
        </p>
      </div>
    );
  }

  return (
    <div className={classNames.wrapper}>
      <h1 className={classNames.title}>{__("Unexpected error :(")}</h1>
      <p className={classNames.description}>{error.toString()}</p>
    </div>
  );
}
