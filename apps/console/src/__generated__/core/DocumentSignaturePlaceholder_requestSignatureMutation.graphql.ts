/**
 * @generated SignedSource<<617f61c937eec7639e3c1c25884f1911>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentVersionSignatureState = "REQUESTED" | "SIGNED";
export type RequestSignatureInput = {
  documentVersionId: string;
  signatoryId: string;
};
export type DocumentSignaturePlaceholder_requestSignatureMutation$variables = {
  connections: ReadonlyArray<string>;
  input: RequestSignatureInput;
};
export type DocumentSignaturePlaceholder_requestSignatureMutation$data = {
  readonly requestSignature: {
    readonly documentVersionSignatureEdge: {
      readonly node: {
        readonly id: string;
        readonly signedBy: {
          readonly id: string;
        };
        readonly state: DocumentVersionSignatureState;
      };
    };
  };
};
export type DocumentSignaturePlaceholder_requestSignatureMutation = {
  response: DocumentSignaturePlaceholder_requestSignatureMutation$data;
  variables: DocumentSignaturePlaceholder_requestSignatureMutation$variables;
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
  "concreteType": "DocumentVersionSignatureEdge",
  "kind": "LinkedField",
  "name": "documentVersionSignatureEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "DocumentVersionSignature",
      "kind": "LinkedField",
      "name": "node",
      "plural": false,
      "selections": [
        (v3/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "state",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "concreteType": "Profile",
          "kind": "LinkedField",
          "name": "signedBy",
          "plural": false,
          "selections": [
            (v3/*: any*/)
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
    "name": "DocumentSignaturePlaceholder_requestSignatureMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "RequestSignaturePayload",
        "kind": "LinkedField",
        "name": "requestSignature",
        "plural": false,
        "selections": [
          (v4/*: any*/)
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
    "name": "DocumentSignaturePlaceholder_requestSignatureMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "RequestSignaturePayload",
        "kind": "LinkedField",
        "name": "requestSignature",
        "plural": false,
        "selections": [
          (v4/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "documentVersionSignatureEdge",
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
    "cacheID": "e1c767d2441f84b783ed4d3824e2f0e8",
    "id": null,
    "metadata": {},
    "name": "DocumentSignaturePlaceholder_requestSignatureMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentSignaturePlaceholder_requestSignatureMutation(\n  $input: RequestSignatureInput!\n) {\n  requestSignature(input: $input) {\n    documentVersionSignatureEdge {\n      node {\n        id\n        state\n        signedBy {\n          id\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "9a5d20aef93c6bf4b83f88851a13b9c9";

export default node;
