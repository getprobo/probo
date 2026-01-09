/**
 * @generated SignedSource<<ef22c5c717241729685db94c9e9c7701>>
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
export type TrustCenterReferenceGraphUpdateRankMutation$variables = {
  input: UpdateTrustCenterReferenceInput;
};
export type TrustCenterReferenceGraphUpdateRankMutation$data = {
  readonly updateTrustCenterReference: {
    readonly trustCenterReference: {
      readonly id: string;
      readonly rank: number;
    };
  };
};
export type TrustCenterReferenceGraphUpdateRankMutation = {
  response: TrustCenterReferenceGraphUpdateRankMutation$data;
  variables: TrustCenterReferenceGraphUpdateRankMutation$variables;
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
            "name": "rank",
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
    "name": "TrustCenterReferenceGraphUpdateRankMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterReferenceGraphUpdateRankMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "9e3d186eff2bff3d66482ab88f27aead",
    "id": null,
    "metadata": {},
    "name": "TrustCenterReferenceGraphUpdateRankMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterReferenceGraphUpdateRankMutation(\n  $input: UpdateTrustCenterReferenceInput!\n) {\n  updateTrustCenterReference(input: $input) {\n    trustCenterReference {\n      id\n      rank\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "72d23bac1e404137890be27bf98bc59e";

export default node;
