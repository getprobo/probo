/**
 * @generated SignedSource<<95a8ba47251f02695b1498614bfd7253>>
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
export type CreateStateOfApplicabilityDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateStateOfApplicabilityInput;
};
export type CreateStateOfApplicabilityDialogMutation$data = {
  readonly createStateOfApplicability: {
    readonly stateOfApplicabilityEdge: {
      readonly node: {
        readonly createdAt: string;
        readonly id: string;
        readonly name: string;
        readonly updatedAt: string;
      };
    };
  };
};
export type CreateStateOfApplicabilityDialogMutation = {
  response: CreateStateOfApplicabilityDialogMutation$data;
  variables: CreateStateOfApplicabilityDialogMutation$variables;
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
    "name": "CreateStateOfApplicabilityDialogMutation",
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
    "name": "CreateStateOfApplicabilityDialogMutation",
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
    "cacheID": "7240fb866d961906a8d6b21b5ee0bf53",
    "id": null,
    "metadata": {},
    "name": "CreateStateOfApplicabilityDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreateStateOfApplicabilityDialogMutation(\n  $input: CreateStateOfApplicabilityInput!\n) {\n  createStateOfApplicability(input: $input) {\n    stateOfApplicabilityEdge {\n      node {\n        id\n        name\n        createdAt\n        updatedAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "f83d337f0110a5f121a885ab110c38fc";

export default node;
