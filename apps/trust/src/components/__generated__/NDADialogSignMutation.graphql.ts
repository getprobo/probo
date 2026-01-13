/**
 * @generated SignedSource<<3ef589b34a5cb80406c517d6e5eb938c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type NDADialogSignMutation$variables = Record<PropertyKey, never>;
export type NDADialogSignMutation$data = {
  readonly acceptNonDisclosureAgreement: {
    readonly success: boolean;
  };
};
export type NDADialogSignMutation = {
  response: NDADialogSignMutation$data;
  variables: NDADialogSignMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "AcceptNonDisclosureAgreementPayload",
    "kind": "LinkedField",
    "name": "acceptNonDisclosureAgreement",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "success",
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "NDADialogSignMutation",
    "selections": (v0/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "NDADialogSignMutation",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "bca0dcb3f227c89d8339b0238aa2e42a",
    "id": null,
    "metadata": {},
    "name": "NDADialogSignMutation",
    "operationKind": "mutation",
    "text": "mutation NDADialogSignMutation {\n  acceptNonDisclosureAgreement {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "814a623444a446730f260e6b7bbf0083";

export default node;
