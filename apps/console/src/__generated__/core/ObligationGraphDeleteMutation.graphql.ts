/**
 * @generated SignedSource<<0966beeae022283847f20f63c3303ee2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteObligationInput = {
  obligationId: string;
};
export type ObligationGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteObligationInput;
};
export type ObligationGraphDeleteMutation$data = {
  readonly deleteObligation: {
    readonly deletedObligationId: string;
  };
};
export type ObligationGraphDeleteMutation = {
  response: ObligationGraphDeleteMutation$data;
  variables: ObligationGraphDeleteMutation$variables;
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
  "name": "deletedObligationId",
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
    "name": "ObligationGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteObligationPayload",
        "kind": "LinkedField",
        "name": "deleteObligation",
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
    "name": "ObligationGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteObligationPayload",
        "kind": "LinkedField",
        "name": "deleteObligation",
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
            "name": "deletedObligationId",
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
    "cacheID": "d54295be0ffd5722cf2e5408a62be33d",
    "id": null,
    "metadata": {},
    "name": "ObligationGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation ObligationGraphDeleteMutation(\n  $input: DeleteObligationInput!\n) {\n  deleteObligation(input: $input) {\n    deletedObligationId\n  }\n}\n"
  }
};
})();

(node as any).hash = "224562b4db5e07b2a11aebda97cd0732";

export default node;
