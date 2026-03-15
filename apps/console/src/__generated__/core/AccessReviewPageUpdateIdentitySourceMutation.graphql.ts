/**
 * @generated SignedSource<<e3cf82d6ebbfe8fcc65bda9119af7808>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateAccessReviewInput = {
  accessReviewId: string;
  identitySourceId?: string | null | undefined;
};
export type AccessReviewPageUpdateIdentitySourceMutation$variables = {
  input: UpdateAccessReviewInput;
};
export type AccessReviewPageUpdateIdentitySourceMutation$data = {
  readonly updateAccessReview: {
    readonly accessReview: {
      readonly id: string;
      readonly identitySource: {
        readonly id: string;
      } | null | undefined;
    };
  };
};
export type AccessReviewPageUpdateIdentitySourceMutation = {
  response: AccessReviewPageUpdateIdentitySourceMutation$data;
  variables: AccessReviewPageUpdateIdentitySourceMutation$variables;
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
    "concreteType": "UpdateAccessReviewPayload",
    "kind": "LinkedField",
    "name": "updateAccessReview",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "AccessReview",
        "kind": "LinkedField",
        "name": "accessReview",
        "plural": false,
        "selections": [
          (v1/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "AccessSource",
            "kind": "LinkedField",
            "name": "identitySource",
            "plural": false,
            "selections": [
              (v1/*: any*/)
            ],
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
    "name": "AccessReviewPageUpdateIdentitySourceMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AccessReviewPageUpdateIdentitySourceMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "88ea9b0c57117672cb1572a0f1c5742c",
    "id": null,
    "metadata": {},
    "name": "AccessReviewPageUpdateIdentitySourceMutation",
    "operationKind": "mutation",
    "text": "mutation AccessReviewPageUpdateIdentitySourceMutation(\n  $input: UpdateAccessReviewInput!\n) {\n  updateAccessReview(input: $input) {\n    accessReview {\n      id\n      identitySource {\n        id\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "df481807ae592c6871d9f37e214b6d7c";

export default node;
