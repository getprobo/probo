/**
 * @generated SignedSource<<aa1a5069891de3be9e8ea2dc554747d4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreatePersonalAPIKeyInput = {
  expiresAt: string;
  name: string;
};
export type PersonalAPIKeyListCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreatePersonalAPIKeyInput;
};
export type PersonalAPIKeyListCreateMutation$data = {
  readonly createPersonalAPIKey: {
    readonly personalAPIKeyEdge: {
      readonly node: {
        readonly createdAt: string;
        readonly expiresAt: string;
        readonly id: string;
        readonly name: string;
      };
    };
    readonly token: string;
  } | null | undefined;
};
export type PersonalAPIKeyListCreateMutation = {
  response: PersonalAPIKeyListCreateMutation$data;
  variables: PersonalAPIKeyListCreateMutation$variables;
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
  "concreteType": "PersonalAPIKeyEdge",
  "kind": "LinkedField",
  "name": "personalAPIKeyEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "PersonalAPIKey",
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
          "name": "expiresAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "token",
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
    "name": "PersonalAPIKeyListCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreatePersonalAPIKeyPayload",
        "kind": "LinkedField",
        "name": "createPersonalAPIKey",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          (v4/*: any*/)
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
    "name": "PersonalAPIKeyListCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreatePersonalAPIKeyPayload",
        "kind": "LinkedField",
        "name": "createPersonalAPIKey",
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
            "name": "personalAPIKeyEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          },
          (v4/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "1ec09abe9f3a61e13ba2deacf1a273fb",
    "id": null,
    "metadata": {},
    "name": "PersonalAPIKeyListCreateMutation",
    "operationKind": "mutation",
    "text": "mutation PersonalAPIKeyListCreateMutation(\n  $input: CreatePersonalAPIKeyInput!\n) {\n  createPersonalAPIKey(input: $input) {\n    personalAPIKeyEdge {\n      node {\n        id\n        name\n        createdAt\n        expiresAt\n      }\n    }\n    token\n  }\n}\n"
  }
};
})();

(node as any).hash = "86df62e5e5f2d9f347f41648d7af9c32";

export default node;
