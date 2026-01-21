/**
 * @generated SignedSource<<56cf69173dcd665f19c4273ab108f383>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteStateOfApplicabilityInput = {
  stateOfApplicabilityId: string;
};
export type StateOfApplicabilityDetailPageDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteStateOfApplicabilityInput;
};
export type StateOfApplicabilityDetailPageDeleteMutation$data = {
  readonly deleteStateOfApplicability: {
    readonly deletedStateOfApplicabilityId: string;
  };
};
export type StateOfApplicabilityDetailPageDeleteMutation = {
  response: StateOfApplicabilityDetailPageDeleteMutation$data;
  variables: StateOfApplicabilityDetailPageDeleteMutation$variables;
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
  "name": "deletedStateOfApplicabilityId",
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
    "name": "StateOfApplicabilityDetailPageDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteStateOfApplicabilityPayload",
        "kind": "LinkedField",
        "name": "deleteStateOfApplicability",
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
    "name": "StateOfApplicabilityDetailPageDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteStateOfApplicabilityPayload",
        "kind": "LinkedField",
        "name": "deleteStateOfApplicability",
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
            "name": "deletedStateOfApplicabilityId",
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
    "cacheID": "13f5a325289c902aca142880e1f30648",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityDetailPageDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation StateOfApplicabilityDetailPageDeleteMutation(\n  $input: DeleteStateOfApplicabilityInput!\n) {\n  deleteStateOfApplicability(input: $input) {\n    deletedStateOfApplicabilityId\n  }\n}\n"
  }
};
})();

(node as any).hash = "b511cc3d55c6ebb1517687f706026333";

export default node;
