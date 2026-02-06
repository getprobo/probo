/**
 * @generated SignedSource<<8fa12501ce4c5a9013357bfa3ee4ff72>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AssetType = "PHYSICAL" | "VIRTUAL";
export type UpdateAssetInput = {
  amount?: number | null | undefined;
  assetType?: AssetType | null | undefined;
  dataTypesStored?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
  ownerId?: string | null | undefined;
  vendorIds?: ReadonlyArray<string> | null | undefined;
};
export type AssetGraphUpdateMutation$variables = {
  input: UpdateAssetInput;
};
export type AssetGraphUpdateMutation$data = {
  readonly updateAsset: {
    readonly asset: {
      readonly amount: number;
      readonly assetType: AssetType;
      readonly dataTypesStored: string;
      readonly id: string;
      readonly name: string;
      readonly owner: {
        readonly fullName: string;
        readonly id: string;
      };
      readonly snapshotId: string | null | undefined;
      readonly updatedAt: string;
      readonly vendors: {
        readonly edges: ReadonlyArray<{
          readonly node: {
            readonly id: string;
            readonly name: string;
            readonly websiteUrl: string | null | undefined;
          };
        }>;
      };
    };
  };
};
export type AssetGraphUpdateMutation = {
  response: AssetGraphUpdateMutation$data;
  variables: AssetGraphUpdateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v3 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateAssetPayload",
    "kind": "LinkedField",
    "name": "updateAsset",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Asset",
        "kind": "LinkedField",
        "name": "asset",
        "plural": false,
        "selections": [
          (v1/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "snapshotId",
            "storageKey": null
          },
          (v2/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "amount",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "assetType",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "dataTypesStored",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "Profile",
            "kind": "LinkedField",
            "name": "owner",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "fullName",
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": [
              {
                "kind": "Literal",
                "name": "first",
                "value": 50
              }
            ],
            "concreteType": "VendorConnection",
            "kind": "LinkedField",
            "name": "vendors",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Vendor",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v1/*: any*/),
                      (v2/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "websiteUrl",
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "vendors(first:50)"
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
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "AssetGraphUpdateMutation",
    "selections": (v3/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AssetGraphUpdateMutation",
    "selections": (v3/*: any*/)
  },
  "params": {
    "cacheID": "ee67fc797e7a09d39db357b1b8e68d11",
    "id": null,
    "metadata": {},
    "name": "AssetGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation AssetGraphUpdateMutation(\n  $input: UpdateAssetInput!\n) {\n  updateAsset(input: $input) {\n    asset {\n      id\n      snapshotId\n      name\n      amount\n      assetType\n      dataTypesStored\n      owner {\n        id\n        fullName\n      }\n      vendors(first: 50) {\n        edges {\n          node {\n            id\n            name\n            websiteUrl\n          }\n        }\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "65a76cfab2b2fba398cd19e566e3ee33";

export default node;
