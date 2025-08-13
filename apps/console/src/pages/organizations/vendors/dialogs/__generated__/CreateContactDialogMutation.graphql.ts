/**
 * @generated SignedSource<<cfc4a09b412cbfa5d046952e1e612107>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateVendorContactInput = {
  email?: string | null | undefined;
  name?: string | null | undefined;
  phone?: string | null | undefined;
  role?: string | null | undefined;
  vendorId: string;
};
export type CreateContactDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateVendorContactInput;
};
export type CreateContactDialogMutation$data = {
  readonly createVendorContact: {
    readonly vendorContactEdge: {
      readonly node: {
        readonly " $fragmentSpreads": FragmentRefs<"VendorContactsTabFragment_contact">;
      };
    };
  };
};
export type CreateContactDialogMutation = {
  response: CreateContactDialogMutation$data;
  variables: CreateContactDialogMutation$variables;
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
    "name": "CreateContactDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateVendorContactPayload",
        "kind": "LinkedField",
        "name": "createVendorContact",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "VendorContactEdge",
            "kind": "LinkedField",
            "name": "vendorContactEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorContact",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "VendorContactsTabFragment_contact"
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
    "name": "CreateContactDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateVendorContactPayload",
        "kind": "LinkedField",
        "name": "createVendorContact",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "VendorContactEdge",
            "kind": "LinkedField",
            "name": "vendorContactEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorContact",
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
                    "name": "email",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "phone",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "role",
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
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "vendorContactEdge",
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
    "cacheID": "df86e6a01a2005491c79821db7d71e65",
    "id": null,
    "metadata": {},
    "name": "CreateContactDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreateContactDialogMutation(\n  $input: CreateVendorContactInput!\n) {\n  createVendorContact(input: $input) {\n    vendorContactEdge {\n      node {\n        ...VendorContactsTabFragment_contact\n        id\n      }\n    }\n  }\n}\n\nfragment VendorContactsTabFragment_contact on VendorContact {\n  id\n  name\n  email\n  phone\n  role\n  createdAt\n  updatedAt\n}\n"
  }
};
})();

(node as any).hash = "f44f76bce42470683c55db50426d3f7b";

export default node;
