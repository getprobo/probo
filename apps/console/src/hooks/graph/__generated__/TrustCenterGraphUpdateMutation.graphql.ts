/**
 * @generated SignedSource<<c257950f58c4b876a6ff7ef6083f005b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateTrustCenterInput = {
  active?: boolean | null | undefined;
  slug?: string | null | undefined;
  trustCenterId: string;
};
export type TrustCenterGraphUpdateMutation$variables = {
  input: UpdateTrustCenterInput;
};
export type TrustCenterGraphUpdateMutation$data = {
  readonly updateTrustCenter: {
    readonly trustCenter: {
      readonly active: boolean;
      readonly id: string;
      readonly slug: string;
      readonly updatedAt: any;
    };
  };
};
export type TrustCenterGraphUpdateMutation = {
  response: TrustCenterGraphUpdateMutation$data;
  variables: TrustCenterGraphUpdateMutation$variables;
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
    "concreteType": "UpdateTrustCenterPayload",
    "kind": "LinkedField",
    "name": "updateTrustCenter",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenter",
        "kind": "LinkedField",
        "name": "trustCenter",
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
            "name": "active",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "slug",
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
    "name": "TrustCenterGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "18e80937923185923d9b427819334337",
    "id": null,
    "metadata": {},
    "name": "TrustCenterGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterGraphUpdateMutation(\n  $input: UpdateTrustCenterInput!\n) {\n  updateTrustCenter(input: $input) {\n    trustCenter {\n      id\n      active\n      slug\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "988f3baf354e729d6cc9482129040898";

export default node;
