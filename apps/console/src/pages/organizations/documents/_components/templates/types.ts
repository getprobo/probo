import type { ComponentType } from "react";

export type DocumentTemplate = {
  id: string;
  icon: ComponentType<{ size?: number; className?: string }>;
  name: string;
  documentType: "OTHER" | "GOVERNANCE" | "POLICY" | "PROCEDURE" | "PLAN" | "REGISTER" | "RECORD" | "REPORT" | "TEMPLATE";
  classification: "PUBLIC" | "INTERNAL" | "CONFIDENTIAL" | "SECRET";
  title: string;
  content: string;
};
