/**
 * @generated SignedSource<<5b430cf79f9527c247a0262d039f9603>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteNonconformityInput = {
  nonconformityId: string;
};
export type NonconformityGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteNonconformityInput;
};
export type NonconformityGraphDeleteMutation$data = {
  readonly deleteNonconformity: {
    readonly deletedNonconformityId: string;
  };
};
export type NonconformityGraphDeleteMutation = {
  response: NonconformityGraphDeleteMutation$data;
  variables: NonconformityGraphDeleteMutation$variables;
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
  "name": "deletedNonconformityId",
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
    "name": "NonconformityGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteNonconformityPayload",
        "kind": "LinkedField",
        "name": "deleteNonconformity",
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
    "name": "NonconformityGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteNonconformityPayload",
        "kind": "LinkedField",
        "name": "deleteNonconformity",
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
            "name": "deletedNonconformityId",
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
    "cacheID": "2e421cc7eb59884d8eeb7b8856f36142",
    "id": null,
    "metadata": {},
    "name": "NonconformityGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation NonconformityGraphDeleteMutation(\n  $input: DeleteNonconformityInput!\n) {\n  deleteNonconformity(input: $input) {\n    deletedNonconformityId\n  }\n}\n"
  }
};
})();

(node as any).hash = "d9d2303a4235301061e976771043d18a";

export default node;
