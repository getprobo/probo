/**
 * @generated SignedSource<<ed7ba35de62460b7625419b78b67997b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteControlInput = {
  controlId: string;
};
export type FrameworkDetailPageDeleteControlMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlInput;
};
export type FrameworkDetailPageDeleteControlMutation$data = {
  readonly deleteControl: {
    readonly deletedControlId: string;
  };
};
export type FrameworkDetailPageDeleteControlMutation = {
  response: FrameworkDetailPageDeleteControlMutation$data;
  variables: FrameworkDetailPageDeleteControlMutation$variables;
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
  "name": "deletedControlId",
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
    "name": "FrameworkDetailPageDeleteControlMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlPayload",
        "kind": "LinkedField",
        "name": "deleteControl",
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
    "name": "FrameworkDetailPageDeleteControlMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlPayload",
        "kind": "LinkedField",
        "name": "deleteControl",
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
            "name": "deletedControlId",
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
    "cacheID": "d86c7a2bd535a98d1cb7308a96c13401",
    "id": null,
    "metadata": {},
    "name": "FrameworkDetailPageDeleteControlMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkDetailPageDeleteControlMutation(\n  $input: DeleteControlInput!\n) {\n  deleteControl(input: $input) {\n    deletedControlId\n  }\n}\n"
  }
};
})();

(node as any).hash = "a067e712c3c51b73abac50e6c3779141";

export default node;
