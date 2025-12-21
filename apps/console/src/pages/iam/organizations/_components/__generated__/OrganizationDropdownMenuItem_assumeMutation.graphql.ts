/**
 * @generated SignedSource<<469707d0eca0a0378fa589809db964bc>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ReauthenticationReason = "POLICY_REQUIREMENT" | "SENSITIVE_ACTION" | "SESSION_EXPIRED";
export type AssumeOrganizationSessionInput = {
  organizationId: string;
};
export type OrganizationDropdownMenuItem_assumeMutation$variables = {
  input: AssumeOrganizationSessionInput;
};
export type OrganizationDropdownMenuItem_assumeMutation$data = {
  readonly assumeOrganizationSession: {
    readonly result: {
      readonly __typename: "PasswordRequired";
      readonly reason: ReauthenticationReason;
    } | {
      readonly __typename: "SAMLAuthenticationRequired";
      readonly reason: ReauthenticationReason;
      readonly redirectUrl: string;
    } | {
      // This will never be '%other', but we need some
      // value in case none of the concrete values match.
      readonly __typename: "%other";
    };
  } | null | undefined;
};
export type OrganizationDropdownMenuItem_assumeMutation = {
  response: OrganizationDropdownMenuItem_assumeMutation$data;
  variables: OrganizationDropdownMenuItem_assumeMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "reason",
  "storageKey": null
},
v2 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "AssumeOrganizationSessionPayload",
    "kind": "LinkedField",
    "name": "assumeOrganizationSession",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": null,
        "kind": "LinkedField",
        "name": "result",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          {
            "kind": "InlineFragment",
            "selections": [
              (v1/*: any*/)
            ],
            "type": "PasswordRequired",
            "abstractKey": null
          },
          {
            "kind": "InlineFragment",
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "redirectUrl",
                "storageKey": null
              }
            ],
            "type": "SAMLAuthenticationRequired",
            "abstractKey": null
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
    "name": "OrganizationDropdownMenuItem_assumeMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "OrganizationDropdownMenuItem_assumeMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "024d06c3bba88bb18eb1aa7e2888edd0",
    "id": null,
    "metadata": {},
    "name": "OrganizationDropdownMenuItem_assumeMutation",
    "operationKind": "mutation",
    "text": "mutation OrganizationDropdownMenuItem_assumeMutation(\n  $input: AssumeOrganizationSessionInput!\n) {\n  assumeOrganizationSession(input: $input) {\n    result {\n      __typename\n      ... on PasswordRequired {\n        reason\n      }\n      ... on SAMLAuthenticationRequired {\n        reason\n        redirectUrl\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "bbc8443e4a756b37f8b46ee98972bdcb";

export default node;
