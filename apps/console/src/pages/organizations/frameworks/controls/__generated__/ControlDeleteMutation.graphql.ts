/**
 * @generated SignedSource<<210832f4f838924747194254c58b57a4>>
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
export type ControlDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlInput;
};
export type ControlDeleteMutation$data = {
  readonly deleteControl: {
    readonly deletedControlId: string;
  };
};
export type ControlDeleteMutation = {
  response: ControlDeleteMutation$data;
  variables: ControlDeleteMutation$variables;
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
    "name": "ControlDeleteMutation",
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
    "name": "ControlDeleteMutation",
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
    "cacheID": "2d95c1ff9328afa01b0f24d1910898d1",
    "id": null,
    "metadata": {},
    "name": "ControlDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation ControlDeleteMutation(\n  $input: DeleteControlInput!\n) {\n  deleteControl(input: $input) {\n    deletedControlId\n  }\n}\n"
  }
};
})();

(node as any).hash = "a1e347241b41977cdea4a31d371e8770";

export default node;
