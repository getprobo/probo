/**
 * @generated SignedSource<<22971bf0238cf9c5805c47c382480109>>
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
  position: string;
  primaryEmailAddress: string;
};
export type CreatePeoplePageCreatePeopleMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreatePeopleInput;
};
export type CreatePeoplePageCreatePeopleMutation$data = {
  readonly createPeople: {
    readonly peopleEdge: {
      readonly node: {
        readonly additionalEmailAddresses: ReadonlyArray<string>;
        readonly fullName: string;
        readonly id: string;
        readonly kind: PeopleKind;
        readonly position: string;
        readonly primaryEmailAddress: string;
      };
    };
  };
};
export type CreatePeoplePageCreatePeopleMutation = {
  response: CreatePeoplePageCreatePeopleMutation$data;
  variables: CreatePeoplePageCreatePeopleMutation$variables;
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
          "name": "additionalEmailAddresses",
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
          "name": "position",
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
    "name": "CreatePeoplePageCreatePeopleMutation",
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
    "name": "CreatePeoplePageCreatePeopleMutation",
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
    "cacheID": "afdb3bc6d7a5080e40cee97baea4efaa",
    "id": null,
    "metadata": {},
    "name": "CreatePeoplePageCreatePeopleMutation",
    "operationKind": "mutation",
    "text": "mutation CreatePeoplePageCreatePeopleMutation(\n  $input: CreatePeopleInput!\n) {\n  createPeople(input: $input) {\n    peopleEdge {\n      node {\n        id\n        fullName\n        primaryEmailAddress\n        additionalEmailAddresses\n        kind\n        position\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "526a7c5eadb073288337fb0f83a45374";

export default node;
