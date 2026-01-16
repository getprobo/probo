/**
 * @generated SignedSource<<deb85307c2d7da7ec52f99a67ad56d9f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteVendorServiceInput = {
  vendorServiceId: string;
};
export type VendorServicesTabDeleteServiceMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteVendorServiceInput;
};
export type VendorServicesTabDeleteServiceMutation$data = {
  readonly deleteVendorService: {
    readonly deletedVendorServiceId: string;
  };
};
export type VendorServicesTabDeleteServiceMutation = {
  response: VendorServicesTabDeleteServiceMutation$data;
  variables: VendorServicesTabDeleteServiceMutation$variables;
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
  "name": "deletedVendorServiceId",
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
    "name": "VendorServicesTabDeleteServiceMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteVendorServicePayload",
        "kind": "LinkedField",
        "name": "deleteVendorService",
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
    "name": "VendorServicesTabDeleteServiceMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteVendorServicePayload",
        "kind": "LinkedField",
        "name": "deleteVendorService",
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
            "name": "deletedVendorServiceId",
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
    "cacheID": "fef00196f74b15a5528a0974541f1ec4",
    "id": null,
    "metadata": {},
    "name": "VendorServicesTabDeleteServiceMutation",
    "operationKind": "mutation",
    "text": "mutation VendorServicesTabDeleteServiceMutation(\n  $input: DeleteVendorServiceInput!\n) {\n  deleteVendorService(input: $input) {\n    deletedVendorServiceId\n  }\n}\n"
  }
};
})();

(node as any).hash = "73d14c22b95d031b2c07b18dbbeb64d4";

export default node;
