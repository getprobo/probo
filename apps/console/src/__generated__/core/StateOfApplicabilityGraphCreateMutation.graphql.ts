/**
 * @generated SignedSource<<3a91be997803d4a28d0db40aebe53f25>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateStateOfApplicabilityInput = {
  name: string;
  organizationId: string;
  ownerId: string;
};
export type StateOfApplicabilityGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateStateOfApplicabilityInput;
};
export type StateOfApplicabilityGraphCreateMutation$data = {
  readonly createStateOfApplicability: {
    readonly stateOfApplicabilityEdge: {
      readonly node: {
        readonly createdAt: any;
        readonly id: string;
        readonly name: string;
        readonly snapshotId: string | null | undefined;
        readonly sourceId: string | null | undefined;
        readonly updatedAt: any;
      };
    };
  };
};
export type StateOfApplicabilityGraphCreateMutation = {
  response: StateOfApplicabilityGraphCreateMutation$data;
  variables: StateOfApplicabilityGraphCreateMutation$variables;
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
  "concreteType": "StateOfApplicabilityEdge",
  "kind": "LinkedField",
  "name": "stateOfApplicabilityEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "StateOfApplicability",
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
          "name": "sourceId",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "snapshotId",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "createdAt",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "updatedAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
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
    "name": "StateOfApplicabilityGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateStateOfApplicabilityPayload",
        "kind": "LinkedField",
        "name": "createStateOfApplicability",
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
    "name": "StateOfApplicabilityGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateStateOfApplicabilityPayload",
        "kind": "LinkedField",
        "name": "createStateOfApplicability",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "stateOfApplicabilityEdge",
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
    "cacheID": "c1a3721f591bc17ed455a326eb9edead",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation StateOfApplicabilityGraphCreateMutation(\n  $input: CreateStateOfApplicabilityInput!\n) {\n  createStateOfApplicability(input: $input) {\n    stateOfApplicabilityEdge {\n      node {\n        id\n        name\n        sourceId\n        snapshotId\n        createdAt\n        updatedAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "294d4fe89f8cbbbb4b8170a46dd3cd7a";

export default node;
