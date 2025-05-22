/**
 * @generated SignedSource<<6b924850fd9bad48c4fe32c331a7770b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteVendorInput = {
  vendorId: string;
};
export type VendorGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteVendorInput;
};
export type VendorGraphDeleteMutation$data = {
  readonly deleteVendor: {
    readonly deletedVendorId: string;
  };
};
export type VendorGraphDeleteMutation = {
  response: VendorGraphDeleteMutation$data;
  variables: VendorGraphDeleteMutation$variables;
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
  "name": "deletedVendorId",
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
    "name": "VendorGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteVendorPayload",
        "kind": "LinkedField",
        "name": "deleteVendor",
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
    "name": "VendorGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteVendorPayload",
        "kind": "LinkedField",
        "name": "deleteVendor",
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
            "name": "deletedVendorId",
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
    "cacheID": "866254027729839312bc72c942de3111",
    "id": null,
    "metadata": {},
    "name": "VendorGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation VendorGraphDeleteMutation(\n  $input: DeleteVendorInput!\n) {\n  deleteVendor(input: $input) {\n    deletedVendorId\n  }\n}\n"
  }
};
})();

(node as any).hash = "82648f6825e035f9abe617d5f641d537";

export default node;
