/**
 * @generated SignedSource<<e4010bea87ff3e2826e9f6eb336e6ffc>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateComplianceBadgeInput = {
  iconFile?: any | null | undefined;
  id: string;
  name?: string | null | undefined;
  rank?: number | null | undefined;
};
export type ComplianceBadgeGraphUpdateMutation$variables = {
  input: UpdateComplianceBadgeInput;
};
export type ComplianceBadgeGraphUpdateMutation$data = {
  readonly updateComplianceBadge: {
    readonly complianceBadge: {
      readonly canDelete: boolean;
      readonly canUpdate: boolean;
      readonly createdAt: string;
      readonly iconUrl: string;
      readonly id: string;
      readonly name: string;
      readonly rank: number;
      readonly updatedAt: string;
    };
  };
};
export type ComplianceBadgeGraphUpdateMutation = {
  response: ComplianceBadgeGraphUpdateMutation$data;
  variables: ComplianceBadgeGraphUpdateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateComplianceBadgePayload",
    "kind": "LinkedField",
    "name": "updateComplianceBadge",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ComplianceBadge",
        "kind": "LinkedField",
        "name": "complianceBadge",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "name",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "iconUrl",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "rank",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "createdAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
            "storageKey": null
          },
          {
            "alias": "canUpdate",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:compliance-badge:update"
              }
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": "permission(action:\"core:compliance-badge:update\")"
          },
          {
            "alias": "canDelete",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:compliance-badge:delete"
              }
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": "permission(action:\"core:compliance-badge:delete\")"
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ComplianceBadgeGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ComplianceBadgeGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "303b5f1aca21778c1e0f17dd59a26fcd",
    "id": null,
    "metadata": {},
    "name": "ComplianceBadgeGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation ComplianceBadgeGraphUpdateMutation(\n  $input: UpdateComplianceBadgeInput!\n) {\n  updateComplianceBadge(input: $input) {\n    complianceBadge {\n      id\n      name\n      iconUrl\n      rank\n      createdAt\n      updatedAt\n      canUpdate: permission(action: \"core:compliance-badge:update\")\n      canDelete: permission(action: \"core:compliance-badge:delete\")\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2ae43f4c1521af8dc1d6e67d1b8928b9";

export default node;
