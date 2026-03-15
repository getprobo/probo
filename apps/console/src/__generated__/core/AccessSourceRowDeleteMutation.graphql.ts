/**
 * @generated SignedSource<<08bd52735e1e6ef59944b80fe3c19ef3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteAccessSourceInput = {
  accessSourceId: string;
};
export type AccessSourceRowDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteAccessSourceInput;
};
export type AccessSourceRowDeleteMutation$data = {
  readonly deleteAccessSource: {
    readonly deletedAccessSourceId: string;
  };
};
export type AccessSourceRowDeleteMutation = {
  response: AccessSourceRowDeleteMutation$data;
  variables: AccessSourceRowDeleteMutation$variables;
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
  "name": "deletedAccessSourceId",
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
    "name": "AccessSourceRowDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteAccessSourcePayload",
        "kind": "LinkedField",
        "name": "deleteAccessSource",
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
    "name": "AccessSourceRowDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteAccessSourcePayload",
        "kind": "LinkedField",
        "name": "deleteAccessSource",
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
            "name": "deletedAccessSourceId",
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
    "cacheID": "29c6154acde3aba5191c3ece1ec991e8",
    "id": null,
    "metadata": {},
    "name": "AccessSourceRowDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation AccessSourceRowDeleteMutation(\n  $input: DeleteAccessSourceInput!\n) {\n  deleteAccessSource(input: $input) {\n    deletedAccessSourceId\n  }\n}\n"
  }
};
})();

(node as any).hash = "130ed30bda5622bf55b9c98592c87379";

export default node;
