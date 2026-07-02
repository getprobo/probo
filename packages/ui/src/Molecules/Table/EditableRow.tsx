// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
