/**
 * @generated SignedSource<<aa74d7b1c9c34a3780220590119aa78f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PeopleKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
export type CreatePeopleInput = {
  additionalEmailAddresses?: ReadonlyArray<string> | null | undefined;
  fullName: string;
  kind: PeopleKind;
  organizationId: string;
  primaryEmailAddress: string;
};
export type CreatePeopleDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreatePeopleInput;
};
export type CreatePeopleDialogMutation$data = {
  readonly createPeople: {
    readonly peopleEdge: {
      readonly node: {
        readonly additionalEmailAddresses: ReadonlyArray<string>;
        readonly fullName: string;
        readonly id: string;
        readonly kind: PeopleKind;
        readonly primaryEmailAddress: string;
      };
    };
  };
};
export type CreatePeopleDialogMutation = {
  response: CreatePeopleDialogMutation$data;
  variables: CreatePeopleDialogMutation$variables;
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
  "concreteType": "PeopleEdge",
  "kind": "LinkedField",
  "name": "peopleEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "People",
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
          "name": "fullName",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "primaryEmailAddress",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "kind",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "additionalEmailAddresses",
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
    "name": "CreatePeopleDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreatePeoplePayload",
        "kind": "LinkedField",
        "name": "createPeople",
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
    "name": "CreatePeopleDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreatePeoplePayload",
        "kind": "LinkedField",
        "name": "createPeople",
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
            "name": "peopleEdge",
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
    "cacheID": "634c2845825a02a0343c3211b41c90a0",
    "id": null,
    "metadata": {},
    "name": "CreatePeopleDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreatePeopleDialogMutation(\n  $input: CreatePeopleInput!\n) {\n  createPeople(input: $input) {\n    peopleEdge {\n      node {\n        id\n        fullName\n        primaryEmailAddress\n        kind\n        additionalEmailAddresses\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "867aad1276c26bd787ddf80c6f166abe";

export default node;
