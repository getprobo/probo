/**
 * @generated SignedSource<<ed4ab6c79b848843012fc091a7eb325c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteTrustCenterAccessInput = {
  id: string;
};
export type TrustCenterAccessGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteTrustCenterAccessInput;
};
export type TrustCenterAccessGraphDeleteMutation$data = {
  readonly deleteTrustCenterAccess: {
    readonly deletedTrustCenterAccessId: string;
  };
};
export type TrustCenterAccessGraphDeleteMutation = {
  response: TrustCenterAccessGraphDeleteMutation$data;
  variables: TrustCenterAccessGraphDeleteMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deletedTrustCenterAccessId",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustCenterAccessGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteTrustCenterAccessPayload",
        "kind": "LinkedField",
        "name": "deleteTrustCenterAccess",
        "plural": false,
        "selections": [
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "TrustCenterAccessGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteTrustCenterAccessPayload",
        "kind": "LinkedField",
        "name": "deleteTrustCenterAccess",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedTrustCenterAccessId",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "3f55c9ce5cac2874b769ec994977367a",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAccessGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterAccessGraphDeleteMutation(\n  $input: DeleteTrustCenterAccessInput!\n) {\n  deleteTrustCenterAccess(input: $input) {\n    deletedTrustCenterAccessId\n  }\n}\n"
  }
};
})();

(node as any).hash = "0d0a44eb6db912eeb533d24718c49b35";

export default node;
