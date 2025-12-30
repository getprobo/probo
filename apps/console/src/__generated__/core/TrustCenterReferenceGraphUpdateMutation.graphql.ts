/**
 * @generated SignedSource<<2b51f7fdbc055d912be03031b258c042>>
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
      readonly createdAt: any;
      readonly description: string | null | undefined;
      readonly id: string;
      readonly logoUrl: string;
      readonly name: string;
      readonly rank: number;
      readonly updatedAt: any;
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
    "cacheID": "f55510d6d1e0a686d4733f5bd82a605a",
    "id": null,
    "metadata": {},
    "name": "TrustCenterReferenceGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterReferenceGraphUpdateMutation(\n  $input: UpdateTrustCenterReferenceInput!\n) {\n  updateTrustCenterReference(input: $input) {\n    trustCenterReference {\n      id\n      name\n      description\n      websiteUrl\n      logoUrl\n      rank\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "bed2a3190570cd85d85dd38f20f375da";

export default node;
