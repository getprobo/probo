/**
 * @generated SignedSource<<d4c60256aa53236cfb4c3a01f4fd35c1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteControlSnapshotMappingInput = {
  controlId: string;
  snapshotId: string;
};
export type FrameworkControlPageDetachSnapshotMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlSnapshotMappingInput;
};
export type FrameworkControlPageDetachSnapshotMutation$data = {
  readonly deleteControlSnapshotMapping: {
    readonly deletedSnapshotId: string;
  };
};
export type FrameworkControlPageDetachSnapshotMutation = {
  response: FrameworkControlPageDetachSnapshotMutation$data;
  variables: FrameworkControlPageDetachSnapshotMutation$variables;
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
  "name": "deletedSnapshotId",
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
    "name": "FrameworkControlPageDetachSnapshotMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlSnapshotMappingPayload",
        "kind": "LinkedField",
        "name": "deleteControlSnapshotMapping",
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
    "name": "FrameworkControlPageDetachSnapshotMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlSnapshotMappingPayload",
        "kind": "LinkedField",
        "name": "deleteControlSnapshotMapping",
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
            "name": "deletedSnapshotId",
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
    "cacheID": "610636745f2c0441c54bd0028fc52cf0",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageDetachSnapshotMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageDetachSnapshotMutation(\n  $input: DeleteControlSnapshotMappingInput!\n) {\n  deleteControlSnapshotMapping(input: $input) {\n    deletedSnapshotId\n  }\n}\n"
  }
};
})();

(node as any).hash = "be8d40beadabe60d8c9bee4d99b885b7";

export default node;
