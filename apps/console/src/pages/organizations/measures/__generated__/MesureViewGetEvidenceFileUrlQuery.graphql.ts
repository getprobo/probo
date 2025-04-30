/**
 * @generated SignedSource<<d3565063e8e1b68e794fe4347c923138>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MeasureViewGetEvidenceFileUrlQuery$variables = {
  evidenceId: string;
};
export type MeasureViewGetEvidenceFileUrlQuery$data = {
  readonly node: {
    readonly fileUrl?: string | null | undefined;
    readonly id?: string;
  };
};
export type MeasureViewGetEvidenceFileUrlQuery = {
  response: MeasureViewGetEvidenceFileUrlQuery$data;
  variables: MeasureViewGetEvidenceFileUrlQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "evidenceId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "evidenceId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fileUrl",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "MeasureViewGetEvidenceFileUrlQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/)
            ],
            "type": "Evidence",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MeasureViewGetEvidenceFileUrlQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/)
            ],
            "type": "Evidence",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "05edf89cc90a90e1a96bd02f8448b629",
    "id": null,
    "metadata": {},
    "name": "MeasureViewGetEvidenceFileUrlQuery",
    "operationKind": "query",
    "text": "query MeasureViewGetEvidenceFileUrlQuery(\n  $evidenceId: ID!\n) {\n  node(id: $evidenceId) {\n    __typename\n    ... on Evidence {\n      id\n      fileUrl\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "2bef4c934e725a40e5140a812bd31bc8";

export default node;
