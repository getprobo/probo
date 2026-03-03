/**
 * @generated SignedSource<<aae03fb419e3ef00b45151ceeb0988cd>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentAccessStatus = "GRANTED" | "REJECTED" | "REQUESTED" | "REVOKED";
export type RequestDocumentAccessInput = {
  documentId: string;
};
export type useRequestAccessCallback_documentMutation$variables = {
  input: RequestDocumentAccessInput;
};
export type useRequestAccessCallback_documentMutation$data = {
  readonly requestDocumentAccess: {
    readonly document: {
      readonly access: {
        readonly id: string;
        readonly status: DocumentAccessStatus;
      } | null | undefined;
    } | null | undefined;
  };
};
export type useRequestAccessCallback_documentMutation = {
  response: useRequestAccessCallback_documentMutation$data;
  variables: useRequestAccessCallback_documentMutation$variables;
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
  "concreteType": "DocumentAccess",
  "kind": "LinkedField",
  "name": "access",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "status",
      "storageKey": null
    }
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "useRequestAccessCallback_documentMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "RequestDocumentAccessPayload",
        "kind": "LinkedField",
        "name": "requestDocumentAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Document",
            "kind": "LinkedField",
            "name": "document",
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
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "useRequestAccessCallback_documentMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "RequestDocumentAccessPayload",
        "kind": "LinkedField",
        "name": "requestDocumentAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Document",
            "kind": "LinkedField",
            "name": "document",
            "plural": false,
            "selections": [
              (v3/*: any*/),
              (v2/*: any*/)
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "e3153093fd3da3f95ac5bc98a09163a7",
    "id": null,
    "metadata": {},
    "name": "useRequestAccessCallback_documentMutation",
    "operationKind": "mutation",
    "text": "mutation useRequestAccessCallback_documentMutation(\n  $input: RequestDocumentAccessInput!\n) {\n  requestDocumentAccess(input: $input) {\n    document {\n      access {\n        id\n        status\n      }\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "afc533a5d5b764f81f6bc7632f1614ff";

export default node;
