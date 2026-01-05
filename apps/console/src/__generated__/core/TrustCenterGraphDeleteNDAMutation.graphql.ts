/**
 * @generated SignedSource<<3220ac89d535b383bbb99661dc296183>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteTrustCenterNDAInput = {
  trustCenterId: string;
};
export type TrustCenterGraphDeleteNDAMutation$variables = {
  input: DeleteTrustCenterNDAInput;
};
export type TrustCenterGraphDeleteNDAMutation$data = {
  readonly deleteTrustCenterNDA: {
    readonly trustCenter: {
      readonly id: string;
      readonly ndaFileName: string | null | undefined;
      readonly updatedAt: string;
    };
  };
};
export type TrustCenterGraphDeleteNDAMutation = {
  response: TrustCenterGraphDeleteNDAMutation$data;
  variables: TrustCenterGraphDeleteNDAMutation$variables;
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
    "concreteType": "DeleteTrustCenterNDAPayload",
    "kind": "LinkedField",
    "name": "deleteTrustCenterNDA",
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
            "name": "ndaFileName",
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
    "name": "TrustCenterGraphDeleteNDAMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterGraphDeleteNDAMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "fa3ebf3593c396bf05819d357d9f909f",
    "id": null,
    "metadata": {},
    "name": "TrustCenterGraphDeleteNDAMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterGraphDeleteNDAMutation(\n  $input: DeleteTrustCenterNDAInput!\n) {\n  deleteTrustCenterNDA(input: $input) {\n    trustCenter {\n      id\n      ndaFileName\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "9c055069b9e2e7432fed7509f6578a07";

export default node;
