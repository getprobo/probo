/**
 * @generated SignedSource<<37dd4db659813ffa6fab65ffec643844>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateVendorInput = {
  businessOwnerId?: string | null | undefined;
  category?: string | null | undefined;
  certifications?: ReadonlyArray<string> | null | undefined;
  dataProcessingAgreementUrl?: string | null | undefined;
  description?: string | null | undefined;
  headquarterAddress?: string | null | undefined;
  legalName?: string | null | undefined;
  name: string;
  organizationId: string;
  privacyPolicyUrl?: string | null | undefined;
  securityOwnerId?: string | null | undefined;
  securityPageUrl?: string | null | undefined;
  serviceLevelAgreementUrl?: string | null | undefined;
  statusPageUrl?: string | null | undefined;
  termsOfServiceUrl?: string | null | undefined;
  trustPageUrl?: string | null | undefined;
  websiteUrl?: string | null | undefined;
};
export type VendorGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateVendorInput;
};
export type VendorGraphCreateMutation$data = {
  readonly createVendor: {
    readonly vendorEdge: {
      readonly node: {
        readonly createdAt: any;
        readonly description: string | null | undefined;
        readonly id: string;
        readonly name: string;
        readonly updatedAt: any;
        readonly websiteUrl: string | null | undefined;
      };
    };
  };
};
export type VendorGraphCreateMutation = {
  response: VendorGraphCreateMutation$data;
  variables: VendorGraphCreateMutation$variables;
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
  "concreteType": "VendorEdge",
  "kind": "LinkedField",
  "name": "vendorEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Vendor",
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
          "name": "websiteUrl",
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
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "VendorGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateVendorPayload",
        "kind": "LinkedField",
        "name": "createVendor",
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
    "name": "VendorGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateVendorPayload",
        "kind": "LinkedField",
        "name": "createVendor",
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
            "name": "vendorEdge",
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
    "cacheID": "b6e686d8c3e4add7f11f09fd0d1e1de3",
    "id": null,
    "metadata": {},
    "name": "VendorGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation VendorGraphCreateMutation(\n  $input: CreateVendorInput!\n) {\n  createVendor(input: $input) {\n    vendorEdge {\n      node {\n        id\n        name\n        description\n        websiteUrl\n        createdAt\n        updatedAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ae71994587599f61caf6b79d9159ad44";

export default node;
