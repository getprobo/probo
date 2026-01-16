/**
 * @generated SignedSource<<22d97ce5cead5a1cc6c23e89fca72cb1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AcceptNonDisclosureAgreementInput = {
  fullName: string;
};
export type NDADialogSignMutation$variables = {
  input: AcceptNonDisclosureAgreementInput;
};
export type NDADialogSignMutation$data = {
  readonly acceptNonDisclosureAgreement: {
    readonly success: boolean;
  } | null | undefined;
};
export type NDADialogSignMutation = {
  response: NDADialogSignMutation$data;
  variables: NDADialogSignMutation$variables;
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
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "NDADialogSignMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "NDADialogSignMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "130cfc307dca0525194e0103a0548bd0",
    "id": null,
    "metadata": {},
    "name": "NDADialogSignMutation",
    "operationKind": "mutation",
    "text": "mutation NDADialogSignMutation(\n  $input: AcceptNonDisclosureAgreementInput!\n) {\n  acceptNonDisclosureAgreement(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "1b9447e5cbb2ec7dce4f3f9c68493555";

export default node;
