/**
 * @generated SignedSource<<e62c28d53365ca5d5f3d53282aefd4e7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RequestDocumentAccessInput = {
  documentId: string;
};
export type DocumentRow_requestAccessMutation$variables = {
  input: RequestDocumentAccessInput;
};
export type DocumentRow_requestAccessMutation$data = {
  readonly requestDocumentAccess: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type DocumentRow_requestAccessMutation = {
  response: DocumentRow_requestAccessMutation$data;
  variables: DocumentRow_requestAccessMutation$variables;
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
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "RequestAccessesPayload",
    "kind": "LinkedField",
    "name": "requestDocumentAccess",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenterAccess",
        "kind": "LinkedField",
        "name": "trustCenterAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
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
    "name": "DocumentRow_requestAccessMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentRow_requestAccessMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "13715b7cb8784d050b6d0381e460bbad",
    "id": null,
    "metadata": {},
    "name": "DocumentRow_requestAccessMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentRow_requestAccessMutation(\n  $input: RequestDocumentAccessInput!\n) {\n  requestDocumentAccess(input: $input) {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ca9026b04c2de91067759e3a80e5c276";

export default node;
