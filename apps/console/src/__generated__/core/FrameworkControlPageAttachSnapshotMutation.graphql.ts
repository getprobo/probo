/**
 * @generated SignedSource<<a4b20010a96dac4779e4b1576b947e45>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateControlSnapshotMappingInput = {
  controlId: string;
  snapshotId: string;
};
export type FrameworkControlPageAttachSnapshotMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateControlSnapshotMappingInput;
};
export type FrameworkControlPageAttachSnapshotMutation$data = {
  readonly createControlSnapshotMapping: {
    readonly snapshotEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"LinkedSnapshotsCardFragment">;
      };
    };
  };
};
export type FrameworkControlPageAttachSnapshotMutation = {
  response: FrameworkControlPageAttachSnapshotMutation$data;
  variables: FrameworkControlPageAttachSnapshotMutation$variables;
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
  "name": "id",
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
    "name": "FrameworkControlPageAttachSnapshotMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlSnapshotMappingPayload",
        "kind": "LinkedField",
        "name": "createControlSnapshotMapping",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "SnapshotEdge",
            "kind": "LinkedField",
            "name": "snapshotEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Snapshot",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "LinkedSnapshotsCardFragment"
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
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
    "name": "FrameworkControlPageAttachSnapshotMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlSnapshotMappingPayload",
        "kind": "LinkedField",
        "name": "createControlSnapshotMapping",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "SnapshotEdge",
            "kind": "LinkedField",
            "name": "snapshotEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Snapshot",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "name",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "description",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "type",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "createdAt",
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "snapshotEdge",
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
    "cacheID": "52c76e336bb7f38dc2d2f6620817100a",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageAttachSnapshotMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageAttachSnapshotMutation(\n  $input: CreateControlSnapshotMappingInput!\n) {\n  createControlSnapshotMapping(input: $input) {\n    snapshotEdge {\n      node {\n        id\n        ...LinkedSnapshotsCardFragment\n      }\n    }\n  }\n}\n\nfragment LinkedSnapshotsCardFragment on Snapshot {\n  id\n  name\n  description\n  type\n  createdAt\n}\n"
  }
};
})();

(node as any).hash = "9e8ad62cc8d9f3672e1f785c31a91624";

export default node;
