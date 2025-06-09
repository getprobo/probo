/**
 * @generated SignedSource<<3a14264d0b2b45f62b29aaefcf7e2c7f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateControlInput = {
  description: string;
  frameworkId: string;
  name: string;
  sectionTitle: string;
};
export type FrameworkControlDialogCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateControlInput;
};
export type FrameworkControlDialogCreateMutation$data = {
  readonly createControl: {
    readonly controlEdge: {
      readonly node: {
        readonly " $fragmentSpreads": FragmentRefs<"FrameworkControlDialogFragment">;
      };
    };
  };
};
export type FrameworkControlDialogCreateMutation = {
  response: FrameworkControlDialogCreateMutation$data;
  variables: FrameworkControlDialogCreateMutation$variables;
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
];
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "FrameworkControlDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlPayload",
        "kind": "LinkedField",
        "name": "createControl",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ControlEdge",
            "kind": "LinkedField",
            "name": "controlEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Control",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "FrameworkControlDialogFragment"
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
    "name": "FrameworkControlDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlPayload",
        "kind": "LinkedField",
        "name": "createControl",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ControlEdge",
            "kind": "LinkedField",
            "name": "controlEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Control",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "id",
                    "storageKey": null
                  },
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
                    "name": "sectionTitle",
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
            "name": "controlEdge",
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
    "cacheID": "87e5bd454a1adfe2f67ecb006e21e115",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlDialogCreateMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlDialogCreateMutation(\n  $input: CreateControlInput!\n) {\n  createControl(input: $input) {\n    controlEdge {\n      node {\n        ...FrameworkControlDialogFragment\n        id\n      }\n    }\n  }\n}\n\nfragment FrameworkControlDialogFragment on Control {\n  id\n  name\n  description\n  sectionTitle\n}\n"
  }
};
})();

(node as any).hash = "a64e6fd936555c0dd2cb4148cde736b6";

export default node;
