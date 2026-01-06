/**
 * @generated SignedSource<<44e9b1516c9cd0ae8419f326791e09e1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { Result } from "relay-runtime";
export type SSOSignInPageQuery$variables = {
  email: string;
};
export type SSOSignInPageQuery$data = {
  readonly ssoLoginURL: Result<string | null | undefined, unknown>;
};
export type SSOSignInPageQuery = {
  response: SSOSignInPageQuery$data;
  variables: SSOSignInPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "email"
  }
],
v1 = {
  "alias": null,
  "args": [
    {
      "kind": "Variable",
      "name": "email",
      "variableName": "email"
    }
  ],
  "kind": "ScalarField",
  "name": "ssoLoginURL",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "SSOSignInPageQuery",
    "selections": [
      {
        "kind": "CatchField",
        "field": (v1/*: any*/),
        "to": "RESULT"
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SSOSignInPageQuery",
    "selections": [
      (v1/*: any*/)
    ]
  },
  "params": {
    "cacheID": "9b7766bbd9f160dc8bbcf34705641486",
    "id": null,
    "metadata": {},
    "name": "SSOSignInPageQuery",
    "operationKind": "query",
    "text": "query SSOSignInPageQuery(\n  $email: EmailAddr!\n) {\n  ssoLoginURL(email: $email)\n}\n"
  }
};
})();

(node as any).hash = "5749c48e67881c71bbc99a9e1120767f";

export default node;
