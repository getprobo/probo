/**
 * @generated SignedSource<<2bb2430e2a2e063fc21f4f2c7c19910f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateTrustCenterReferenceInput = {
  description?: string | null | undefined;
  id: string;
  logoFile?: any | null | undefined;
  name?: string | null | undefined;
  rank?: number | null | undefined;
  websiteUrl?: string | null | undefined;
};
export type TrustCenterReferenceGraphUpdateMutation$variables = {
  input: UpdateTrustCenterReferenceInput;
};
export type TrustCenterReferenceGraphUpdateMutation$data = {
  readonly updateTrustCenterReference: {
    readonly trustCenterReference: {
      readonly canDelete: boolean;
      readonly canUpdate: boolean;
      readonly createdAt: string;
      readonly description: string | null | undefined;
      readonly id: string;
      readonly logoUrl: string;
      readonly name: string;
      readonly rank: number;
      readonly updatedAt: string;
      readonly websiteUrl: string;
    };
  };
};
export type TrustCenterReferenceGraphUpdateMutation = {
  response: TrustCenterReferenceGraphUpdateMutation$data;
  variables: TrustCenterReferenceGraphUpdateMutation$variables;
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
    "concreteType": "UpdateTrustCenterReferencePayload",
    "kind": "LinkedField",
    "name": "updateTrustCenterReference",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenterReference",
        "kind": "LinkedField",
        "name": "trustCenterReference",
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
            "name": "description",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "websiteUrl",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "logoUrl",
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
                "value": "core:trust-center-reference:update"
              }
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": "permission(action:\"core:trust-center-reference:update\")"
          },
          {
            "alias": "canDelete",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:trust-center-reference:delete"
              }
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": "permission(action:\"core:trust-center-reference:delete\")"
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
    "name": "TrustCenterReferenceGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterReferenceGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "7c0e441a6e51c68b72f3dc0b74079da1",
    "id": null,
    "metadata": {},
    "name": "TrustCenterReferenceGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterReferenceGraphUpdateMutation(\n  $input: UpdateTrustCenterReferenceInput!\n) {\n  updateTrustCenterReference(input: $input) {\n    trustCenterReference {\n      id\n      name\n      description\n      websiteUrl\n      logoUrl\n      rank\n      createdAt\n      updatedAt\n      canUpdate: permission(action: \"core:trust-center-reference:update\")\n      canDelete: permission(action: \"core:trust-center-reference:delete\")\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "f6add85271c2b048654bf586745252fc";

export default node;
