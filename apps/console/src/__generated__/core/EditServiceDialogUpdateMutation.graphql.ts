/**
 * @generated SignedSource<<97f0c6ed319e3fba440079ef085a55eb>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type UpdateVendorServiceInput = {
  description?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
  type?: string | null | undefined;
  url?: string | null | undefined;
};
export type EditServiceDialogUpdateMutation$variables = {
  input: UpdateVendorServiceInput;
};
export type EditServiceDialogUpdateMutation$data = {
  readonly updateVendorService: {
    readonly vendorService: {
      readonly " $fragmentSpreads": FragmentRefs<"VendorServicesTabFragment_service">;
    };
  };
};
export type EditServiceDialogUpdateMutation = {
  response: EditServiceDialogUpdateMutation$data;
  variables: EditServiceDialogUpdateMutation$variables;
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
    "name": "EditServiceDialogUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateVendorServicePayload",
        "kind": "LinkedField",
        "name": "updateVendorService",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "VendorService",
            "kind": "LinkedField",
            "name": "vendorService",
            "plural": false,
            "selections": [
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "VendorServicesTabFragment_service"
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
    "name": "EditServiceDialogUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateVendorServicePayload",
        "kind": "LinkedField",
        "name": "updateVendorService",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "VendorService",
            "kind": "LinkedField",
            "name": "vendorService",
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
                "name": "description",
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
    "cacheID": "f2c37fa9a0b7f0ed043a4529d1f50c65",
    "id": null,
    "metadata": {},
    "name": "EditServiceDialogUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation EditServiceDialogUpdateMutation(\n  $input: UpdateVendorServiceInput!\n) {\n  updateVendorService(input: $input) {\n    vendorService {\n      ...VendorServicesTabFragment_service\n      id\n    }\n  }\n}\n\nfragment VendorServicesTabFragment_service on VendorService {\n  id\n  name\n  description\n  createdAt\n  updatedAt\n}\n"
  }
};
})();

(node as any).hash = "8a34720f5ae7e1f5c619518d4e0db460";

export default node;
