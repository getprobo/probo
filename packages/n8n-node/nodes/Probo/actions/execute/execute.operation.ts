import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Query',
		name: 'query',
		type: 'string',
		typeOptions: {
			rows: 5,
		},
		displayOptions: {
			show: {
				resource: ['execute'],
			},
		},
		default: '',
		description: 'The complete GraphQL operation including operation name and variable declarations (e.g., "query GetUser($userId: ID!) { node(ID: $userId) { ID } }" or "mutation UpdateUser($input: UpdateUserInput!) { updateUser(input: $input) { ID } }")',
		required: true,
	},
	{
		displayName: 'Variables',
		name: 'variables',
		type: 'json',
		displayOptions: {
			show: {
				resource: ['execute'],
			},
		},
		default: '{}',
		description: 'GraphQL variables as JSON object',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const query = this.getNodeParameter('query', itemIndex) as string;
	const variablesParam = this.getNodeParameter('variables', itemIndex) as string;

	// Basic validation: check if query contains a GraphQL operation
	const trimmedQuery = query.trim();
	if (!trimmedQuery) {
		throw new Error('GraphQL query cannot be empty');
	}

	// Check for operation type (query, mutation, or subscription)
	const operationMatch = trimmedQuery.match(/^\s*(query|mutation|subscription)\s+(\w+)/i);
	if (!operationMatch) {
		throw new Error(
			'GraphQL operation must start with "query", "mutation", or "subscription" followed by an operation name (e.g., "query GetUser { ... }" or "mutation UpdateUser { ... }")',
		);
	}

	let variables = {};
	if (variablesParam) {
		try {
			variables =
				typeof variablesParam === 'string' ? JSON.parse(variablesParam) : variablesParam;
		} catch (error) {
			throw new Error(`Invalid JSON in Variables: ${error instanceof Error ? error.message : String(error)}`);
		}
	}

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
