/**
 * @generated SignedSource<<5b93fe65ad69563bbcb6e75d0b913eb8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DataSensitivity = "CRITICAL" | "HIGH" | "LOW" | "MEDIUM" | "NONE";
export type CreateDatumInput = {
  dataSensitivity: DataSensitivity;
  name: string;
  organizationId: string;
  ownerId: string;
  vendorIds?: ReadonlyArray<string> | null | undefined;
};
export type NewDatumViewCreateDatumMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateDatumInput;
};
export type NewDatumViewCreateDatumMutation$data = {
  readonly createDatum: {
    readonly datumEdge: {
      readonly node: {
        readonly dataSensitivity: DataSensitivity;
        readonly id: string;
        readonly name: string;
        readonly owner: {
          readonly fullName: string;
          readonly id: string;
        };
        readonly vendors: {
          readonly edges: ReadonlyArray<{
            readonly node: {
              readonly id: string;
              readonly name: string;
            };
          }>;
        };
      };
    };
  };
};
export type NewDatumViewCreateDatumMutation = {
  response: NewDatumViewCreateDatumMutation$data;
  variables: NewDatumViewCreateDatumMutation$variables;
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
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "concreteType": "DatumEdge",
  "kind": "LinkedField",
  "name": "datumEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Datum",
      "kind": "LinkedField",
      "name": "node",
      "plural": false,
      "selections": [
        (v3/*: any*/),
        (v4/*: any*/),
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
          "concreteType": "People",
          "kind": "LinkedField",
          "name": "owner",
          "plural": false,
          "selections": [
            (v3/*: any*/),
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
                    (v3/*: any*/),
                    (v4/*: any*/)
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
    "name": "NewDatumViewCreateDatumMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateDatumPayload",
        "kind": "LinkedField",
        "name": "createDatum",
        "plural": false,
        "selections": [
          (v5/*: any*/)
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
    "name": "NewDatumViewCreateDatumMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateDatumPayload",
        "kind": "LinkedField",
        "name": "createDatum",
        "plural": false,
        "selections": [
          (v5/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "datumEdge",
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
    "cacheID": "0ab550803b091f5dfa1935f810eaecdc",
    "id": null,
    "metadata": {},
    "name": "NewDatumViewCreateDatumMutation",
    "operationKind": "mutation",
    "text": "mutation NewDatumViewCreateDatumMutation(\n  $input: CreateDatumInput!\n) {\n  createDatum(input: $input) {\n    datumEdge {\n      node {\n        id\n        name\n        dataSensitivity\n        owner {\n          id\n          fullName\n        }\n        vendors {\n          edges {\n            node {\n              id\n              name\n            }\n          }\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "697a260fb7d44115c617682d38b9259d";

export default node;
