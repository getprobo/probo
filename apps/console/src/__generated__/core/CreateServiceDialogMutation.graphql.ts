/**
 * @generated SignedSource<<aacbb3c4646610f327cc7cea8f944158>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateVendorServiceInput = {
  description?: string | null | undefined;
  name: string;
  type?: string | null | undefined;
  url?: string | null | undefined;
  vendorId: string;
};
export type CreateServiceDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateVendorServiceInput;
};
export type CreateServiceDialogMutation$data = {
  readonly createVendorService: {
    readonly vendorServiceEdge: {
      readonly node: {
        readonly " $fragmentSpreads": FragmentRefs<"VendorServicesTabFragment_service">;
      };
    };
  };
};
export type CreateServiceDialogMutation = {
  response: CreateServiceDialogMutation$data;
  variables: CreateServiceDialogMutation$variables;
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
    "name": "CreateServiceDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateVendorServicePayload",
        "kind": "LinkedField",
        "name": "createVendorService",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "VendorServiceEdge",
            "kind": "LinkedField",
            "name": "vendorServiceEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorService",
                "kind": "LinkedField",
                "name": "node",
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
    "name": "CreateServiceDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateVendorServicePayload",
        "kind": "LinkedField",
        "name": "createVendorService",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "VendorServiceEdge",
            "kind": "LinkedField",
            "name": "vendorServiceEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorService",
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
                  },
                  {
                    "alias": "canUpdate",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:vendor-service:update"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:vendor-service:update\")"
                  },
                  {
                    "alias": "canDelete",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:vendor-service:delete"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:vendor-service:delete\")"
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
            "name": "vendorServiceEdge",
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
    "cacheID": "05250ac44a987544335a9e97eed9d66b",
    "id": null,
    "metadata": {},
    "name": "CreateServiceDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreateServiceDialogMutation(\n  $input: CreateVendorServiceInput!\n) {\n  createVendorService(input: $input) {\n    vendorServiceEdge {\n      node {\n        ...VendorServicesTabFragment_service\n        id\n      }\n    }\n  }\n}\n\nfragment VendorServicesTabFragment_service on VendorService {\n  id\n  name\n  description\n  createdAt\n  updatedAt\n  canUpdate: permission(action: \"core:vendor-service:update\")\n  canDelete: permission(action: \"core:vendor-service:delete\")\n}\n"
  }
};
})();

(node as any).hash = "8dd0e535bfa15e8ac78b630139b57177";

export default node;
