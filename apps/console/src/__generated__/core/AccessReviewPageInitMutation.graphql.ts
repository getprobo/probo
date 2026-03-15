/**
 * @generated SignedSource<<df1fc08b2113c21b00fbd88d38b25270>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateAccessReviewInput = {
  organizationId: string;
};
export type AccessReviewPageInitMutation$variables = {
  input: CreateAccessReviewInput;
};
export type AccessReviewPageInitMutation$data = {
  readonly createAccessReview: {
    readonly accessReview: {
      readonly id: string;
    };
  };
};
export type AccessReviewPageInitMutation = {
  response: AccessReviewPageInitMutation$data;
  variables: AccessReviewPageInitMutation$variables;
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
    "concreteType": "CreateAccessReviewPayload",
    "kind": "LinkedField",
    "name": "createAccessReview",
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
    "name": "AccessReviewPageInitMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AccessReviewPageInitMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "ac37aaff87a75f39cf51fe9e1edf5847",
    "id": null,
    "metadata": {},
    "name": "AccessReviewPageInitMutation",
    "operationKind": "mutation",
    "text": "mutation AccessReviewPageInitMutation(\n  $input: CreateAccessReviewInput!\n) {\n  createAccessReview(input: $input) {\n    accessReview {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "878ad63f0995da8a1bdf916622bdd2b6";

export default node;
