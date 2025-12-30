/**
 * @generated SignedSource<<9acf6a96c0398ea5ae6fec887e8d01b1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PeopleKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
export type CreatePeopleInput = {
  additionalEmailAddresses: ReadonlyArray<any>;
  contractEndDate?: any | null | undefined;
  contractStartDate?: any | null | undefined;
  fullName: string;
  kind: PeopleKind;
  organizationId: string;
  position?: string | null | undefined;
  primaryEmailAddress: any;
};
export type CreatePeopleDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreatePeopleInput;
};
export type CreatePeopleDialogMutation$data = {
  readonly createPeople: {
    readonly peopleEdge: {
      readonly node: {
        readonly additionalEmailAddresses: ReadonlyArray<any>;
        readonly fullName: string;
        readonly id: string;
        readonly kind: PeopleKind;
        readonly position: string | null | undefined;
        readonly primaryEmailAddress: any;
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
          "name": "position",
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
    "cacheID": "4065779551f13f5555fa8a0f1b1ae54c",
    "id": null,
    "metadata": {},
    "name": "CreatePeopleDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreatePeopleDialogMutation(\n  $input: CreatePeopleInput!\n) {\n  createPeople(input: $input) {\n    peopleEdge {\n      node {\n        id\n        fullName\n        primaryEmailAddress\n        position\n        kind\n        additionalEmailAddresses\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "1fbca12cb827104dc700c70a677f92aa";

export default node;
