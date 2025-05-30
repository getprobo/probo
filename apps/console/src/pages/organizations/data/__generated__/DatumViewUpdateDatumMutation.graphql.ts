/**
 * @generated SignedSource<<4aa66259b93ace779a3e38430cf8c1ee>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DataSensitivity = "CRITICAL" | "HIGH" | "LOW" | "MEDIUM" | "NONE";
export type UpdateDatumInput = {
  dataSensitivity?: DataSensitivity | null | undefined;
  id: string;
  name?: string | null | undefined;
  ownerId?: string | null | undefined;
  vendorIds?: ReadonlyArray<string> | null | undefined;
};
export type DatumViewUpdateDatumMutation$variables = {
  input: UpdateDatumInput;
};
export type DatumViewUpdateDatumMutation$data = {
  readonly updateDatum: {
    readonly datum: {
      readonly dataSensitivity: DataSensitivity;
      readonly id: string;
      readonly name: string;
      readonly owner: {
        readonly fullName: string;
        readonly id: string;
      };
      readonly updatedAt: string;
      readonly vendors: {
        readonly edges: ReadonlyArray<{
          readonly node: {
            readonly id: string;
          };
        }>;
      };
    };
  };
};
export type DatumViewUpdateDatumMutation = {
  response: DatumViewUpdateDatumMutation$data;
  variables: DatumViewUpdateDatumMutation$variables;
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
v2 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateDatumPayload",
    "kind": "LinkedField",
    "name": "updateDatum",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Datum",
        "kind": "LinkedField",
        "name": "datum",
        "plural": false,
        "selections": [
          (v1/*: any*/),
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
            "name": "dataSensitivity",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
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
                      (v1/*: any*/)
                    ],
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
            "concreteType": "People",
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
    "name": "DatumViewUpdateDatumMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DatumViewUpdateDatumMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "1e5b2489558fef74becf57d2457a4891",
    "id": null,
    "metadata": {},
    "name": "DatumViewUpdateDatumMutation",
    "operationKind": "mutation",
    "text": "mutation DatumViewUpdateDatumMutation(\n  $input: UpdateDatumInput!\n) {\n  updateDatum(input: $input) {\n    datum {\n      id\n      name\n      dataSensitivity\n      vendors {\n        edges {\n          node {\n            id\n          }\n        }\n      }\n      owner {\n        id\n        fullName\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ac445fa9f2ccebef8203e0138c794ea8";

export default node;
