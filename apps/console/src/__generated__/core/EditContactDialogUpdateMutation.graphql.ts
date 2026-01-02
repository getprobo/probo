/**
 * @generated SignedSource<<d22f20fafedcdddc50cd21917a8a5c28>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type UpdateVendorContactInput = {
  email?: string | null | undefined;
  fullName?: string | null | undefined;
  id: string;
  phone?: string | null | undefined;
  role?: string | null | undefined;
};
export type EditContactDialogUpdateMutation$variables = {
  input: UpdateVendorContactInput;
};
export type EditContactDialogUpdateMutation$data = {
  readonly updateVendorContact: {
    readonly vendorContact: {
      readonly " $fragmentSpreads": FragmentRefs<"VendorContactsTabFragment_contact">;
    };
  };
};
export type EditContactDialogUpdateMutation = {
  response: EditContactDialogUpdateMutation$data;
  variables: EditContactDialogUpdateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "EditContactDialogUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateVendorContactPayload",
        "kind": "LinkedField",
        "name": "updateVendorContact",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "VendorContact",
            "kind": "LinkedField",
            "name": "vendorContact",
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
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditContactDialogUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateVendorContactPayload",
        "kind": "LinkedField",
        "name": "updateVendorContact",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "VendorContact",
            "kind": "LinkedField",
            "name": "vendorContact",
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
      }
    ]
  },
  "params": {
    "cacheID": "bacc163156572b5da49705a27084f5c9",
    "id": null,
    "metadata": {},
    "name": "EditContactDialogUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation EditContactDialogUpdateMutation(\n  $input: UpdateVendorContactInput!\n) {\n  updateVendorContact(input: $input) {\n    vendorContact {\n      ...VendorContactsTabFragment_contact\n      id\n    }\n  }\n}\n\nfragment VendorContactsTabFragment_contact on VendorContact {\n  id\n  fullName\n  email\n  phone\n  role\n  createdAt\n  updatedAt\n}\n"
  }
};
})();

(node as any).hash = "845750bb920bd01ed91bbae1531aaaf0";

export default node;
