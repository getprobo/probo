import type { INodeProperties } from 'n8n-workflow';
import * as executeOp from './execute.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['execute'],
			},
		},
		options: [
			{
				name: 'Execute',
				value: 'execute',
				description: 'Execute a GraphQL query or mutation',
				action: 'Execute graphql',
			},
		],
		default: 'execute',
	},
	...executeOp.description,
];

export { executeOp as execute };

