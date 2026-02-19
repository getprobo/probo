/**
 * @generated SignedSource<<1f378775ea75fa89e23e4858bbef38e1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteControlReportMappingInput = {
  controlId: string;
  reportId: string;
};
export type FrameworkControlPageDetachReportMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlReportMappingInput;
};
export type FrameworkControlPageDetachReportMutation$data = {
  readonly deleteControlReportMapping: {
    readonly deletedReportId: string;
  };
};
export type FrameworkControlPageDetachReportMutation = {
  response: FrameworkControlPageDetachReportMutation$data;
  variables: FrameworkControlPageDetachReportMutation$variables;
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
  "name": "deletedReportId",
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
    "name": "FrameworkControlPageDetachReportMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlReportMappingPayload",
        "kind": "LinkedField",
        "name": "deleteControlReportMapping",
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
    "name": "FrameworkControlPageDetachReportMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlReportMappingPayload",
        "kind": "LinkedField",
        "name": "deleteControlReportMapping",
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
            "name": "deletedReportId",
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
    "cacheID": "60d21b84842ae23f7a0950edda1d0077",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageDetachReportMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageDetachReportMutation(\n  $input: DeleteControlReportMappingInput!\n) {\n  deleteControlReportMapping(input: $input) {\n    deletedReportId\n  }\n}\n"
  }
};
})();

(node as any).hash = "c862b08e6b34ca90c5189df6d4dd682b";

export default node;
