/**
 * @generated SignedSource<<87a74ac3a3c2523bd65512703dc2246e>>
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
export type FrameworkControlPageDeleteControlMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlInput;
};
export type FrameworkControlPageDeleteControlMutation$data = {
  readonly deleteControl: {
    readonly deletedControlId: string;
  };
};
export type FrameworkControlPageDeleteControlMutation = {
  response: FrameworkControlPageDeleteControlMutation$data;
  variables: FrameworkControlPageDeleteControlMutation$variables;
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
    "name": "FrameworkControlPageDeleteControlMutation",
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
    "name": "FrameworkControlPageDeleteControlMutation",
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
    "cacheID": "f5559c3bf33da22ba6cb83fc408426e0",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageDeleteControlMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageDeleteControlMutation(\n  $input: DeleteControlInput!\n) {\n  deleteControl(input: $input) {\n    deletedControlId\n  }\n}\n"
  }
};
})();

(node as any).hash = "c92c75d0666d2a60d3a2fcfb2959b2e9";

export default node;
