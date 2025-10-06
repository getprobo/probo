import { useLocation, useRouteError } from "react-router";
import { IconPageCross } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useEffect, useRef } from "react";

const classNames = {
  wrapper: "py-10 text-center space-y-2 ",
  title: "text-2xl flex gap-2 font-semibold items-center justify-center",
  description: "text-base text-txt-tertiary",
  detail:
    "text-sm text-txt-tertiary font-mono text-start border border-border-low p-2 rounded bg-level-1 mt-2",
};

type Props = {
  resetErrorBoundary?: () => void;
  error?: string;
};

export function PageError({ resetErrorBoundary, error: propsError }: Props) {
  const error = useRouteError() ?? propsError;
  const { __ } = useTranslate();
  const location = useLocation();
  const baseLocation = useRef(location);

  // Reset error boundary on page change
  useEffect(() => {
    if (
      location.pathname !== baseLocation.current.pathname &&
      resetErrorBoundary
    ) {
      resetErrorBoundary();
    }
  }, [location, resetErrorBoundary]);

  if (!error || (error && error.toString().includes("PAGE_NOT_FOUND"))) {
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

  if (error && error.toString().includes("UNAUTHORIZED")) {
    return (
      <div className={classNames.wrapper}>
        <h1 className={classNames.title}>
          <IconPageCross size={26} />
          {__("Access denied")}
        </h1>
        <p className={classNames.description}>
          {__("You don't have permission to access this organization")}
        </p>
      </div>
    );
  }

  return (
    <div className={classNames.wrapper}>
      <h1 className={classNames.title}>{__("Unexpected error :(")}</h1>
      <details>
        <summary className={classNames.description}>
          {__("Something went wrong")}
        </summary>
        <p className={classNames.detail}>{error.toString()}</p>
      </details>
    </div>
  );
}
