/**
 * @generated SignedSource<<b5a89cb2acee801ab7d41cf2afb471b6>>
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
export type useAssumeMutation$variables = {
  input: AssumeOrganizationSessionInput;
};
export type useAssumeMutation$data = {
  readonly assumeOrganizationSession: {
    readonly result: {
      readonly __typename: "OrganizationSessionCreated";
      readonly membership: {
        readonly id: string;
        readonly lastSession: {
          readonly expiresAt: string;
          readonly id: string;
        } | null | undefined;
      };
    } | {
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
export type useAssumeMutation = {
  response: useAssumeMutation$data;
  variables: useAssumeMutation$variables;
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
  "name": "id",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "reason",
  "storageKey": null
},
v3 = [
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
              {
                "alias": null,
                "args": null,
                "concreteType": "Membership",
                "kind": "LinkedField",
                "name": "membership",
                "plural": false,
                "selections": [
                  (v1/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Session",
                    "kind": "LinkedField",
                    "name": "lastSession",
                    "plural": false,
                    "selections": [
                      (v1/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "expiresAt",
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "type": "OrganizationSessionCreated",
            "abstractKey": null
          },
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/)
            ],
            "type": "PasswordRequired",
            "abstractKey": null
          },
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
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
    "name": "useAssumeMutation",
    "selections": (v3/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "useAssumeMutation",
    "selections": (v3/*: any*/)
  },
  "params": {
    "cacheID": "6c65712730108428976b1d0007ca2993",
    "id": null,
    "metadata": {},
    "name": "useAssumeMutation",
    "operationKind": "mutation",
    "text": "mutation useAssumeMutation(\n  $input: AssumeOrganizationSessionInput!\n) {\n  assumeOrganizationSession(input: $input) {\n    result {\n      __typename\n      ... on OrganizationSessionCreated {\n        membership {\n          id\n          lastSession {\n            id\n            expiresAt\n          }\n        }\n      }\n      ... on PasswordRequired {\n        reason\n      }\n      ... on SAMLAuthenticationRequired {\n        reason\n        redirectUrl\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "1ccd539abfc276084939857e65421be9";

export default node;
