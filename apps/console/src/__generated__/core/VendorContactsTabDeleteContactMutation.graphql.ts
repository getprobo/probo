/**
 * @generated SignedSource<<3824bce11d22a60d1e50f53c699611c5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteVendorContactInput = {
  vendorContactId: string;
};
export type VendorContactsTabDeleteContactMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteVendorContactInput;
};
export type VendorContactsTabDeleteContactMutation$data = {
  readonly deleteVendorContact: {
    readonly deletedVendorContactId: string;
  };
};
export type VendorContactsTabDeleteContactMutation = {
  response: VendorContactsTabDeleteContactMutation$data;
  variables: VendorContactsTabDeleteContactMutation$variables;
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
  "name": "deletedVendorContactId",
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
    "name": "VendorContactsTabDeleteContactMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteVendorContactPayload",
        "kind": "LinkedField",
        "name": "deleteVendorContact",
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
    "name": "VendorContactsTabDeleteContactMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteVendorContactPayload",
        "kind": "LinkedField",
        "name": "deleteVendorContact",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedVendorContactId",
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
    "cacheID": "806c0b076e6da2ac4841d172047c8655",
    "id": null,
    "metadata": {},
    "name": "VendorContactsTabDeleteContactMutation",
    "operationKind": "mutation",
    "text": "mutation VendorContactsTabDeleteContactMutation(\n  $input: DeleteVendorContactInput!\n) {\n  deleteVendorContact(input: $input) {\n    deletedVendorContactId\n  }\n}\n"
  }
};
})();

(node as any).hash = "f0595bed2e2e7527eeac8737a0640a4d";

export default node;
