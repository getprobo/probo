/**
 * @generated SignedSource<<72b5471e12ed01f059b4192ee5827a1a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteMailingListSubscriberInput = {
  id: string;
};
export type CompliancePageMailingListDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteMailingListSubscriberInput;
};
export type CompliancePageMailingListDeleteMutation$data = {
  readonly deleteMailingListSubscriber: {
    readonly deletedMailingListSubscriberId: string;
  };
};
export type CompliancePageMailingListDeleteMutation = {
  response: CompliancePageMailingListDeleteMutation$data;
  variables: CompliancePageMailingListDeleteMutation$variables;
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
  "name": "deletedMailingListSubscriberId",
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
    "name": "CompliancePageMailingListDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteMailingListSubscriberPayload",
        "kind": "LinkedField",
        "name": "deleteMailingListSubscriber",
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
    "name": "CompliancePageMailingListDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteMailingListSubscriberPayload",
        "kind": "LinkedField",
        "name": "deleteMailingListSubscriber",
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
            "name": "deletedMailingListSubscriberId",
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
    "cacheID": "65de43d2f00cd2c553aa72568f81eac2",
    "id": null,
    "metadata": {},
    "name": "CompliancePageMailingListDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageMailingListDeleteMutation(\n  $input: DeleteMailingListSubscriberInput!\n) {\n  deleteMailingListSubscriber(input: $input) {\n    deletedMailingListSubscriberId\n  }\n}\n"
  }
};
})();

(node as any).hash = "bb627bfeb68e7b72c337c40133ab63a8";

export default node;
