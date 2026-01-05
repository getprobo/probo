/**
 * @generated SignedSource<<274e16758d468bf7fb911965acca3264>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteRightsRequestInput = {
  rightsRequestId: string;
};
export type RightsRequestGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteRightsRequestInput;
};
export type RightsRequestGraphDeleteMutation$data = {
  readonly deleteRightsRequest: {
    readonly deletedRightsRequestId: string;
  };
};
export type RightsRequestGraphDeleteMutation = {
  response: RightsRequestGraphDeleteMutation$data;
  variables: RightsRequestGraphDeleteMutation$variables;
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
  "name": "deletedRightsRequestId",
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
    "name": "RightsRequestGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteRightsRequestPayload",
        "kind": "LinkedField",
        "name": "deleteRightsRequest",
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
    "name": "RightsRequestGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteRightsRequestPayload",
        "kind": "LinkedField",
        "name": "deleteRightsRequest",
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
            "name": "deletedRightsRequestId",
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
    "cacheID": "c81b44380a4647a241c72fe1ca359246",
    "id": null,
    "metadata": {},
    "name": "RightsRequestGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation RightsRequestGraphDeleteMutation(\n  $input: DeleteRightsRequestInput!\n) {\n  deleteRightsRequest(input: $input) {\n    deletedRightsRequestId\n  }\n}\n"
  }
};
})();

(node as any).hash = "e657b62ad808fa4fd5902cf4eea4ca3b";

export default node;
