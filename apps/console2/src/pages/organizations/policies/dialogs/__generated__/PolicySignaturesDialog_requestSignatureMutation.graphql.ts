/**
 * @generated SignedSource<<af123c927212dd99fbaeb81e3b2f2bd9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type PolicyVersionSignatureState = "REQUESTED" | "SIGNED";
export type RequestSignatureInput = {
  policyVersionId: string;
  signatoryId: string;
};
export type PolicySignaturesDialog_requestSignatureMutation$variables = {
  connections: ReadonlyArray<string>;
  input: RequestSignatureInput;
};
export type PolicySignaturesDialog_requestSignatureMutation$data = {
  readonly requestSignature: {
    readonly policyVersionSignatureEdge: {
      readonly node: {
        readonly id: string;
        readonly signedBy: {
          readonly id: string;
        };
        readonly state: PolicyVersionSignatureState;
        readonly " $fragmentSpreads": FragmentRefs<"PolicySignaturesDialog_signature">;
      };
    };
  };
};
export type PolicySignaturesDialog_requestSignatureMutation = {
  response: PolicySignaturesDialog_requestSignatureMutation$data;
  variables: PolicySignaturesDialog_requestSignatureMutation$variables;
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
    "name": "PolicySignaturesDialog_requestSignatureMutation",
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
            "concreteType": "PolicyVersionSignatureEdge",
            "kind": "LinkedField",
            "name": "policyVersionSignatureEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "PolicyVersionSignature",
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
                    "name": "PolicySignaturesDialog_signature"
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
    "name": "PolicySignaturesDialog_requestSignatureMutation",
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
            "concreteType": "PolicyVersionSignatureEdge",
            "kind": "LinkedField",
            "name": "policyVersionSignatureEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "PolicyVersionSignature",
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
            "name": "policyVersionSignatureEdge",
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
    "cacheID": "20c460adbc3480622c4c91220a6e54fe",
    "id": null,
    "metadata": {},
    "name": "PolicySignaturesDialog_requestSignatureMutation",
    "operationKind": "mutation",
    "text": "mutation PolicySignaturesDialog_requestSignatureMutation(\n  $input: RequestSignatureInput!\n) {\n  requestSignature(input: $input) {\n    policyVersionSignatureEdge {\n      node {\n        id\n        state\n        signedBy {\n          id\n        }\n        ...PolicySignaturesDialog_signature\n      }\n    }\n  }\n}\n\nfragment PolicySignaturesDialog_signature on PolicyVersionSignature {\n  id\n  state\n  signedAt\n  requestedAt\n  signedBy {\n    fullName\n    primaryEmailAddress\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "24825b1c4040debb7edda76c8ad3d662";

export default node;
