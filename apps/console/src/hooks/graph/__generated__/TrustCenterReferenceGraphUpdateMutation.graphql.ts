/**
 * @generated SignedSource<<3b02f7734c467adddc8f8abe1c18c4e5>>
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
  websiteUrl?: string | null | undefined;
};
export type TrustCenterReferenceGraphUpdateMutation$variables = {
  input: UpdateTrustCenterReferenceInput;
};
export type TrustCenterReferenceGraphUpdateMutation$data = {
  readonly updateTrustCenterReference: {
    readonly trustCenterReference: {
      readonly createdAt: any;
      readonly description: string;
      readonly id: string;
      readonly logoUrl: string;
      readonly name: string;
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
    "cacheID": "e5fa3bdb21897523c1491d6cfbf816cf",
    "id": null,
    "metadata": {},
    "name": "TrustCenterReferenceGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterReferenceGraphUpdateMutation(\n  $input: UpdateTrustCenterReferenceInput!\n) {\n  updateTrustCenterReference(input: $input) {\n    trustCenterReference {\n      id\n      name\n      description\n      websiteUrl\n      logoUrl\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2340a9c559d302025df08a5748c18b80";

export default node;
