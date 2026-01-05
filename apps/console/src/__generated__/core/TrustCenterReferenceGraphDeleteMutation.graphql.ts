/**
 * @generated SignedSource<<1f4c8af006c90496e08b285944c83eac>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteTrustCenterReferenceInput = {
  id: string;
};
export type TrustCenterReferenceGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteTrustCenterReferenceInput;
};
export type TrustCenterReferenceGraphDeleteMutation$data = {
  readonly deleteTrustCenterReference: {
    readonly deletedTrustCenterReferenceId: string;
  };
};
export type TrustCenterReferenceGraphDeleteMutation = {
  response: TrustCenterReferenceGraphDeleteMutation$data;
  variables: TrustCenterReferenceGraphDeleteMutation$variables;
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
  "name": "deletedTrustCenterReferenceId",
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
    "name": "TrustCenterReferenceGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteTrustCenterReferencePayload",
        "kind": "LinkedField",
        "name": "deleteTrustCenterReference",
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
    "name": "TrustCenterReferenceGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteTrustCenterReferencePayload",
        "kind": "LinkedField",
        "name": "deleteTrustCenterReference",
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
            "name": "deletedTrustCenterReferenceId",
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
    "cacheID": "6d22b36eb6ce9b568bfc091741a63ea8",
    "id": null,
    "metadata": {},
    "name": "TrustCenterReferenceGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterReferenceGraphDeleteMutation(\n  $input: DeleteTrustCenterReferenceInput!\n) {\n  deleteTrustCenterReference(input: $input) {\n    deletedTrustCenterReferenceId\n  }\n}\n"
  }
};
})();

(node as any).hash = "31a0fcb302daa8f42bc5ecb74fadfd0b";

export default node;
