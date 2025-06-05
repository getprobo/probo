/**
 * @generated SignedSource<<fa604130d8725651770072b638fc5926>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentVersionSignatureState = "REQUESTED" | "SIGNED";
export type RequestSignatureInput = {
  documentVersionId: string;
  signatoryId: string;
};
export type DocumentSignaturesDialog_requestSignatureMutation$variables = {
  connections: ReadonlyArray<string>;
  input: RequestSignatureInput;
};
export type DocumentSignaturesDialog_requestSignatureMutation$data = {
  readonly requestSignature: {
    readonly documentVersionSignatureEdge: {
      readonly node: {
        readonly id: string;
        readonly signedBy: {
          readonly id: string;
        };
        readonly state: DocumentVersionSignatureState;
        readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturesDialog_signature">;
      };
    };
  };
};
export type DocumentSignaturesDialog_requestSignatureMutation = {
  response: DocumentSignaturesDialog_requestSignatureMutation$data;
  variables: DocumentSignaturesDialog_requestSignatureMutation$variables;
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
  "name": "state",
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
    "name": "DocumentSignaturesDialog_requestSignatureMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "RequestSignaturePayload",
        "kind": "LinkedField",
        "name": "requestSignature",
        "plural": false,
        "selections": [
          {
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
                  (v4/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "People",
                    "kind": "LinkedField",
                    "name": "signedBy",
                    "plural": false,
                    "selections": [
                      (v3/*: any*/)
                    ],
                    "storageKey": null
                  },
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "DocumentSignaturesDialog_signature"
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
    "name": "DocumentSignaturesDialog_requestSignatureMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "RequestSignaturePayload",
        "kind": "LinkedField",
        "name": "requestSignature",
        "plural": false,
        "selections": [
          {
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
                  (v4/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "People",
                    "kind": "LinkedField",
                    "name": "signedBy",
                    "plural": false,
                    "selections": [
                      (v3/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "fullName",
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "primaryEmailAddress",
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "signedAt",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "requestedAt",
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
    "cacheID": "6a5413db69a07e0eef99245e949cf4b2",
    "id": null,
    "metadata": {},
    "name": "DocumentSignaturesDialog_requestSignatureMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentSignaturesDialog_requestSignatureMutation(\n  $input: RequestSignatureInput!\n) {\n  requestSignature(input: $input) {\n    documentVersionSignatureEdge {\n      node {\n        id\n        state\n        signedBy {\n          id\n        }\n        ...DocumentSignaturesDialog_signature\n      }\n    }\n  }\n}\n\nfragment DocumentSignaturesDialog_signature on DocumentVersionSignature {\n  id\n  state\n  signedAt\n  requestedAt\n  signedBy {\n    fullName\n    primaryEmailAddress\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "6c51f681411cd3de162c78a3db20ec30";

export default node;
