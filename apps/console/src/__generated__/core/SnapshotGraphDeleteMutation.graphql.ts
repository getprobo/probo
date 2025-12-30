/**
 * @generated SignedSource<<6a696f2d1ebd4a2e733bd9d1b3c42017>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteSnapshotInput = {
  snapshotId: string;
};
export type SnapshotGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteSnapshotInput;
};
export type SnapshotGraphDeleteMutation$data = {
  readonly deleteSnapshot: {
    readonly deletedSnapshotId: string;
  };
};
export type SnapshotGraphDeleteMutation = {
  response: SnapshotGraphDeleteMutation$data;
  variables: SnapshotGraphDeleteMutation$variables;
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
    "name": "SnapshotGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteSnapshotPayload",
        "kind": "LinkedField",
        "name": "deleteSnapshot",
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
    "name": "SnapshotGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteSnapshotPayload",
        "kind": "LinkedField",
        "name": "deleteSnapshot",
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
    "cacheID": "f5f81e50b61dbc8d5668e421cf88223f",
    "id": null,
    "metadata": {},
    "name": "SnapshotGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation SnapshotGraphDeleteMutation(\n  $input: DeleteSnapshotInput!\n) {\n  deleteSnapshot(input: $input) {\n    deletedSnapshotId\n  }\n}\n"
  }
};
})();

(node as any).hash = "cc7cb0955bf74c611136392354ff8dcf";

export default node;
