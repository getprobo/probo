/**
 * @generated SignedSource<<242c546027e5b7c8410fca920a4dfa3c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeletePeopleInput = {
  peopleId: string;
};
export type PeopleGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeletePeopleInput;
};
export type PeopleGraphDeleteMutation$data = {
  readonly deletePeople: {
    readonly deletedPeopleId: string;
  };
};
export type PeopleGraphDeleteMutation = {
  response: PeopleGraphDeleteMutation$data;
  variables: PeopleGraphDeleteMutation$variables;
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
  "name": "deletedPeopleId",
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
    "name": "PeopleGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeletePeoplePayload",
        "kind": "LinkedField",
        "name": "deletePeople",
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
    "name": "PeopleGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeletePeoplePayload",
        "kind": "LinkedField",
        "name": "deletePeople",
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
            "name": "deletedPeopleId",
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
    "cacheID": "9e0f5fd398468199333312505f258bb2",
    "id": null,
    "metadata": {},
    "name": "PeopleGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation PeopleGraphDeleteMutation(\n  $input: DeletePeopleInput!\n) {\n  deletePeople(input: $input) {\n    deletedPeopleId\n  }\n}\n"
  }
};
})();

(node as any).hash = "611aa91bbb039c4e5800838f5db7985c";

export default node;
