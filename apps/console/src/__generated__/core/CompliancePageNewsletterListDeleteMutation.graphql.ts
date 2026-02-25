/**
 * @generated SignedSource<<d019db2bd3820f19cf30608260983723>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteNewsletterSubscriberInput = {
  id: string;
};
export type CompliancePageNewsletterListDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteNewsletterSubscriberInput;
};
export type CompliancePageNewsletterListDeleteMutation$data = {
  readonly deleteNewsletterSubscriber: {
    readonly deletedNewsletterSubscriberId: string;
  };
};
export type CompliancePageNewsletterListDeleteMutation = {
  response: CompliancePageNewsletterListDeleteMutation$data;
  variables: CompliancePageNewsletterListDeleteMutation$variables;
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
  "name": "deletedNewsletterSubscriberId",
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
    "name": "CompliancePageNewsletterListDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteNewsletterSubscriberPayload",
        "kind": "LinkedField",
        "name": "deleteNewsletterSubscriber",
        "plural": false,
        "selections": [
          (v3/*: any*/)
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
    "name": "CompliancePageNewsletterListDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteNewsletterSubscriberPayload",
        "kind": "LinkedField",
        "name": "deleteNewsletterSubscriber",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedNewsletterSubscriberId",
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
    "cacheID": "c02f1499ae102eb3f2b95211cd3918ad",
    "id": null,
    "metadata": {},
    "name": "CompliancePageNewsletterListDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageNewsletterListDeleteMutation(\n  $input: DeleteNewsletterSubscriberInput!\n) {\n  deleteNewsletterSubscriber(input: $input) {\n    deletedNewsletterSubscriberId\n  }\n}\n"
  }
};
})();

(node as any).hash = "2530fdc049352ab4d86e2bd7ee386da8";

export default node;
