import { RouterProvider } from "react-router";
import { router } from "./routes";
import { Toasts } from "@probo/ui";
import { useLang } from "./providers/TranslatorProvider";
import { useTranslate } from "@probo/i18n";

export function App() {
  const { setLang } = useLang();
  const { __ } = useTranslate();

  return (
    <>
      <div
        style={{
          position: "absolute",
          top: 0,
          right: 0,
          padding: "1rem",
          zIndex: 9999,
        }}
      >
        <button onClick={() => setLang("en")}>English</button>
        <button onClick={() => setLang("fr")}>French</button>
      </div>
      <h1>{__("Hello")}</h1>
      <RouterProvider router={router} />
      <Toasts />
    </>
  );
}
