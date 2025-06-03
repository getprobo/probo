/**
 * @generated SignedSource<<863e9f8552e5bb472cc4b02a6dbad1bf>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from "relay-runtime";
export type BusinessImpact = "CRITICAL" | "HIGH" | "LOW" | "MEDIUM";
export type DataSensitivity = "CRITICAL" | "HIGH" | "LOW" | "MEDIUM" | "NONE";
export type CreateVendorRiskAssessmentInput = {
  assessedBy: string;
  businessImpact: BusinessImpact;
  dataSensitivity: DataSensitivity;
  expiresAt: string;
  notes?: string | null | undefined;
  vendorId: string;
};
export type VendorViewCreateRiskAssessmentMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateVendorRiskAssessmentInput;
};
export type VendorViewCreateRiskAssessmentMutation$data = {
  readonly createVendorRiskAssessment: {
    readonly vendorRiskAssessmentEdge: {
      readonly node: {
        readonly assessedAt: string;
        readonly assessedBy: {
          readonly fullName: string;
          readonly id: string;
        };
        readonly businessImpact: BusinessImpact;
        readonly createdAt: string;
        readonly dataSensitivity: DataSensitivity;
        readonly expiresAt: string;
        readonly id: string;
        readonly notes: string | null | undefined;
      };
    };
  };
};
export type VendorViewCreateRiskAssessmentMutation = {
  response: VendorViewCreateRiskAssessmentMutation$data;
  variables: VendorViewCreateRiskAssessmentMutation$variables;
};

const node: ConcreteRequest = (function () {
  var v0 = {
      defaultValue: null,
      kind: "LocalArgument",
      name: "connections",
    },
    v1 = {
      defaultValue: null,
      kind: "LocalArgument",
      name: "input",
    },
    v2 = [
      {
        kind: "Variable",
        name: "input",
        variableName: "input",
      },
    ],
    v3 = {
      alias: null,
      args: null,
      kind: "ScalarField",
      name: "id",
      storageKey: null,
    },
    v4 = {
      alias: null,
      args: null,
      concreteType: "VendorRiskAssessmentEdge",
      kind: "LinkedField",
      name: "vendorRiskAssessmentEdge",
      plural: false,
      selections: [
        {
          alias: null,
          args: null,
          concreteType: "VendorRiskAssessment",
          kind: "LinkedField",
          name: "node",
          plural: false,
          selections: [
            v3 /*: any*/,
            {
              alias: null,
              args: null,
              kind: "ScalarField",
              name: "assessedAt",
              storageKey: null,
            },
            {
              alias: null,
              args: null,
              kind: "ScalarField",
              name: "expiresAt",
              storageKey: null,
            },
            {
              alias: null,
              args: null,
              kind: "ScalarField",
              name: "dataSensitivity",
              storageKey: null,
            },
            {
              alias: null,
              args: null,
              kind: "ScalarField",
              name: "businessImpact",
              storageKey: null,
            },
            {
              alias: null,
              args: null,
              kind: "ScalarField",
              name: "notes",
              storageKey: null,
            },
            {
              alias: null,
              args: null,
              concreteType: "People",
              kind: "LinkedField",
              name: "assessedBy",
              plural: false,
              selections: [
                v3 /*: any*/,
                {
                  alias: null,
                  args: null,
                  kind: "ScalarField",
                  name: "fullName",
                  storageKey: null,
                },
              ],
              storageKey: null,
            },
            {
              alias: null,
              args: null,
              kind: "ScalarField",
              name: "createdAt",
              storageKey: null,
            },
          ],
          storageKey: null,
        },
      ],
      storageKey: null,
    };
  return {
    fragment: {
      argumentDefinitions: [v0 /*: any*/, v1 /*: any*/],
      kind: "Fragment",
      metadata: null,
      name: "VendorViewCreateRiskAssessmentMutation",
      selections: [
        {
          alias: null,
          args: v2 /*: any*/,
          concreteType: "CreateVendorRiskAssessmentPayload",
          kind: "LinkedField",
          name: "createVendorRiskAssessment",
          plural: false,
          selections: [v4 /*: any*/],
          storageKey: null,
        },
      ],
      type: "Mutation",
      abstractKey: null,
    },
    kind: "Request",
    operation: {
      argumentDefinitions: [v1 /*: any*/, v0 /*: any*/],
      kind: "Operation",
      name: "VendorViewCreateRiskAssessmentMutation",
      selections: [
        {
          alias: null,
          args: v2 /*: any*/,
          concreteType: "CreateVendorRiskAssessmentPayload",
          kind: "LinkedField",
          name: "createVendorRiskAssessment",
          plural: false,
          selections: [
            v4 /*: any*/,
            {
              alias: null,
              args: null,
              filters: null,
              handle: "prependEdge",
              key: "",
              kind: "LinkedHandle",
              name: "vendorRiskAssessmentEdge",
              handleArgs: [
                {
                  kind: "Variable",
                  name: "connections",
                  variableName: "connections",
                },
              ],
            },
          ],
          storageKey: null,
        },
      ],
    },
    params: {
      cacheID: "1238320b8db75088917d0e5796652f50",
      id: null,
      metadata: {},
      name: "VendorViewCreateRiskAssessmentMutation",
      operationKind: "mutation",
      text: "mutation VendorViewCreateRiskAssessmentMutation(\n  $input: CreateVendorRiskAssessmentInput!\n) {\n  createVendorRiskAssessment(input: $input) {\n    vendorRiskAssessmentEdge {\n      node {\n        id\n        assessedAt\n        expiresAt\n        dataSensitivity\n        businessImpact\n        notes\n        assessedBy {\n          id\n          fullName\n        }\n        createdAt\n      }\n    }\n  }\n}\n",
    },
  };
})();

(node as any).hash = "d28e057eba5f7287f3d0f5e38d444ca5";

export default node;
