import {
  createContext,
  type ReactNode,
  useContext,
  useMemo,
  useRef,
} from "react";

import { Row } from "../../Atoms/DataTable/DataTable";

type Props = {
  onUpdate: (key: string, value: unknown) => void;
  errors?: Record<string, string>;
  children: ReactNode;
};

export const EditableRowContext = createContext<null | Omit<Props, "children">>(
  null,
);

export const useEditableRowContext = () => {
  const context = useContext(EditableRowContext);
  if (!context) {
    throw new Error(
      "useEditableRowContext must be used within an EditableRow",
    );
  }
  return context;
};

export function EditableRow(props: Props) {
  const onUpdateRef = useRef(props.onUpdate);

  const value = useMemo(
    () => ({
      errors: props.errors,
      onUpdate: (key: string, value: unknown) =>
        onUpdateRef.current(key, value),
    }),
    [props.errors],
  );
  return (
    <EditableRowContext.Provider value={value}>
      <Row>{props.children}</Row>
    </EditableRowContext.Provider>
  );
}
